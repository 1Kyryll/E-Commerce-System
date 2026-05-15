# System Design - Decisions & Trade-offs

This document present each design decision, the reasoning behind it, and the trade-offs faced during the designing process. This system ensures it can handle a real-world scenarios with a great deal of users traffic and a projected workload on the database. System is designed to face and effectively handle such concerns as: eventual consistency, partial failure, concurrency, with explicit and premeditated choices rather than accidental ones. 

#### Each subset is a mini-ADR which presents the final decision, alternatives and the reason behind each one of them.

### 1. Inventory - atomic conditional decrement and row locking

There were couple of approaches to solving this task properly. The goals were: high throughput/performance under heavy load, data consistency, low contention. So I'd come to a conclusion that a single SQL query like this is enough: 

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

### 2. Reservation table for the payment window

We need a Reservations table in order to provide explicit and queryable reservations lifecycle. We rely on the reservations and not the absence of the inventory. Each reservation comes with a *15 minute TTL*.

**Tradeoff:** Clean-up worker for row expirations and an additional table. 

**Alternatives:** 
- *Decrement on payment success.* Fails on 10K users trying to purchase and 10 units of a product available at the moment. Both users see the item is available, after the payment succeeded the product is suddenly out of stock.
- *Decrement on add-to-cart.* Not every product added to cart will be purchased.
- *No Reservations, rely on Order status.* Forces frequent scans on the `orders` table which is exprensive, also comes up with complicated code.

### 3. Idempotency on order creation

Every order-creation query should be idempotent(multiple queries result same as one) and include `Idempotency-Key` header. The key is enforced as a unique constraint on the `reservations` table. A retried request with the same key returns the original order, not a new one.

**Tradeoff:** Client must generate and persist idempotency keys per checkout attempt. 

**Alternatives:** 
- *Trust the client.* Fails on every retry, network blip, or back-button.
- *Deduplicate by user + product + recent window.* Brittle — two intentional purchases of the same item within the window would be silently merged.

### 4. Cart: persistent, not session-bound

According to [System Requirements](https://github.com/1Kyryll/E-Commerce-System/main/docs/adr/001-system-requirements.md), `Cart` must be persistent. Therefore, every add-to-cart request is saved to the `cart` table keyed by `User ID`. 

**Tradeoff:** Cart writes become DB writes (more load), but cart-recovery is free — a user who closes the tab and comes back tomorrow finds their cart. 

**Alternatives:**
- *Browser cookie / localStorage.* User loses their cart on device change or browser clear.
- *Server-side session in Redis.* Lost on Redis eviction or restart.

### 5. Async fan-out via the transactional outbox

When an order is finalized, an event row is written to the `outbox` table in the same transaction as the order. A separate worker polls the table and publishes events to downstream services(email, analytics, recommendations) making each as sent after acknowledgement.

**Tradeoff:** Adds the outbox table and a worker. In exchange, we get exactly-once-or-more event delivery (consumers must be idempotent, which is normal anyway), full audit trail, and the ability to replay events for new consumers.

**Alternatives:**
- *Publish directly to the message broker after DB commit.* Classic dual-write problem: an Order is written to the database, message broker fails, downstream services never get notified.
- *Simutaneously call each service inside the order transaction.* Adds latency, everyone waits until each services finishes it's job. One service slows down the whole checkout process.

### 6. Catalog reads - cache-aside and cursor pagination 

Request for a catalog hits the Redis cache first(name, description, image, stock is left for DB query), if nothing is found, then hit the DB directly. Use cursor pagination for queryinh chunks of data.

**Tradeoff:** if we also want *in-stock* property in cache, the data might be stale for a couple of seconds. However, it isn't a big deal because if the user clicks such an item with slightly stale data, he will immediately get rejected. Cursor pagination might introduce complications related to cursor(uuid, published_at) and DB's data type casting. 

**Alternatives:**
- *Offset pagination.* It is more straightforward than cursor pagination, but that comes with probability of page loss/appearance under constant scrolls or adding/deleting products to DB.
- *No cache.* Every page render hits the DB, which doesn't scale quite well.
- *Cache inventory count.* Inventory count must be consistent, it is constantly changing, what doesn't overlap with caching concept.
