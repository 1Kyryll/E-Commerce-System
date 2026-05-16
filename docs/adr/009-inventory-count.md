## 9. Inventory - total minus active reservations

The conceptual model is straightforward: at any moment, the available inventory of a product equals its physical total minus the quantity reserved by in-flight purchases that haven't yet completed or expired.
 
```
available(product) = inventory_total(product) - Σ active_reservations.quantity
```
 
This is the model that makes the purchase-flow concurrency story work. A "Buy" click creates a reservation, which counts against availability for everyone else; the reservation either becomes a finalized order (consumed) or expires (released), and the count adjusts. The [purchase flow document](../system-design.md) walks through the lifecycle in detail; this section is about the schema and the implementation choices that follow from the conceptual model.
 
The implementation maintains both `inventory_total` (the physical count, changes only on stock receipt) and `inventory_available` (the *materialized* current availability) on the products row. The reservations table tracks each in-flight purchase with a TTL. Atomic decrement against `inventory_available` is what handles the 10K-concurrent-buyers case cheaply.
 
```sql
CREATE TABLE products (
    id                  UUID    PRIMARY KEY,
    name                TEXT    NOT NULL,
    price_amount        NUMERIC(18, 4) NOT NULL,
    price_currency      CHAR(3) NOT NULL,
    inventory_total     INT     NOT NULL CHECK (inventory_total >= 0),
    inventory_available INT     NOT NULL CHECK (inventory_available >= 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
 
CREATE TABLE reservations (
    id              UUID    PRIMARY KEY,
    idempotency_key UUID    NOT NULL UNIQUE,
    product_id      UUID    NOT NULL REFERENCES products(id),
    user_id         UUID    NOT NULL REFERENCES users(id),
    quantity        INT     NOT NULL CHECK (quantity > 0),
    status          TEXT    NOT NULL CHECK (status IN ('active', 'consumed', 'released')),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    consumed_at     TIMESTAMPTZ,
    released_at     TIMESTAMPTZ
);
 
-- Cleanup worker: find expired actives fast
CREATE INDEX idx_reservations_active_expiring
    ON reservations (expires_at) WHERE status = 'active';
 
-- "What's holding this product right now"
CREATE INDEX idx_reservations_active_by_product
    ON reservations (product_id) WHERE status = 'active';
```
 
Two things worth noticing about this schema. First, `inventory_available` exists as its own column — a *denormalization* of `inventory_total - Σ active reservations`, maintained transactionally on every reservation create or release. The obvious alternative is to compute it on every read via JOIN and aggregation, but that forces every product browse to scan the reservations table and requires serializable isolation (or advisory locks) to make the decrement atomic. The materialized version is dramatically faster under contention and cleaner to reason about. The denormalization is justified by the access pattern: availability is read constantly (every page render) and written exactly when reservations transition state.
 
Second, the partial indexes (`WHERE status = 'active'`) cut the index size dramatically. The active set is tiny relative to the all-time set of reservations ever created — most of the table is `consumed` or `released` history. The index only carries the rows that actually need fast lookup.
 
The atomic decrement that creates a reservation is one SQL statement, which Postgres treats as atomic without any explicit transaction:
 
```sql
WITH decremented AS (
    UPDATE products
       SET inventory_available = inventory_available - $1,
           updated_at = now()
     WHERE id = $2
       AND inventory_available >= $1
    RETURNING id
)
INSERT INTO reservations
    (id, idempotency_key, product_id, user_id, quantity, status, expires_at)
SELECT $3, $4, $2, $5, $1, 'active', now() + INTERVAL '15 minutes'
  FROM decremented;
```
 
If the UPDATE matches a row, the decrement succeeds and the CTE produces a row, so the INSERT runs. If `inventory_available < quantity`, the UPDATE matches nothing, the CTE is empty, and the INSERT inserts zero rows — the application sees zero affected rows and translates that to `409 Conflict` for the user. Postgres handles single-row contention via MVCC, so ten thousand concurrent attempts on the same product resolve in microseconds with exactly N successes, where N is the remaining availability.
 
On a successful payment, the reservation is consumed without any inventory change — the decrement already happened at reservation time:
 
```sql
UPDATE reservations
   SET status = 'consumed', consumed_at = now()
 WHERE idempotency_key = $1
   AND status = 'active'
RETURNING id;
```
 
On payment failure or TTL expiry, the reservation is released *and* inventory is restored in the same transaction, so the invariant never breaks:
 
```sql
WITH released AS (
    UPDATE reservations
       SET status = 'released', released_at = now()
     WHERE id = $1
       AND status = 'active'
    RETURNING product_id, quantity
)
UPDATE products p
   SET inventory_available = p.inventory_available + r.quantity
  FROM released r
 WHERE p.id = r.product_id;
```
 
The cleanup worker runs this for every reservation past `expires_at` on its schedule.
 
### The invariant check
 
Because `inventory_available` is a denormalization, it can theoretically drift if a code bug fails to maintain it. A nightly reconciliation job verifies the invariant for every product:
 
```sql
SELECT p.id,
       p.inventory_available AS materialized,
       p.inventory_total - COALESCE(r.reserved, 0) AS computed
  FROM products p
  LEFT JOIN (
       SELECT product_id, SUM(quantity) AS reserved
         FROM reservations
        WHERE status = 'active'
        GROUP BY product_id
  ) r ON r.product_id = p.id
 WHERE p.inventory_available <> p.inventory_total - COALESCE(r.reserved, 0);
```
 
This should always return zero rows. If it ever doesn't, that's an alert — a code bug to investigate, not a number to silently fix. The result of this reconciliation is the single most important durability metric the system has: a system where this drifts is a system that oversells. It runs as a scheduled job and writes its result as a Prometheus gauge (`inventory_invariant_violations_total`), with an alert on any non-zero value.
 