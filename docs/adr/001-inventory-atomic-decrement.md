### 1. Inventory - atomic conditional decrement and row locking

There were couple of approaches to solving this task properly. The goals were: high throughput/performance under heavy load, data consistency, low contention.

**Decision.** So I'd come to a conclusion that a single SQL query like this is enough:

```sql
UPDATE products
SET inventory = inventory - 1
WHERE id = $1 AND inventory > 0
RETURNING id;
```

If `RETURNING` produces a row, the caller got a unit. If it returns nothing, the product is sold out — the caller is rejected immediately.

**Tradeoff:** might include a small amount of contention on a single DB row for hot products. Even so, DB(Postgres) handles that well under the hood. For Black-Friday-scale flash sales, a token-queue pattern in Redis would be needed.

**Alternatives:**
- *Application-level distributed lock (Redis SETNX).* New point of failure(lock holder crashes), overhead, confusing locks handling with timing(TTL), DB already solves that natively.
- *Pessimistic locking* via `SELECT ... FOR UPDATE`. Works correctly and guarantees consistent data which comes with high contention, waiting times for customers, low performance and throughput. This approach would be nice for a banking system, not e-commerce.
- *Optimistic Locking* forces retry loops for a logically single decision. Comes with data inconsistencies and multiple sources of truth(depending on the implementation), meanwhile assures reduced contention and adequate performance.
