### 5. Async fan-out via the transactional outbox

When an order is finalized, an event row is written to the `outbox` table in the same transaction as the order. A separate worker polls the table and publishes events to downstream services(email, analytics, recommendations) making each as sent after acknowledgement.

**Tradeoff:** Adds the outbox table and a worker. In exchange, we get exactly-once-or-more event delivery (consumers must be idempotent, which is normal anyway), full audit trail, and the ability to replay events for new consumers.

**Alternatives:**
- *Publish directly to the message broker after DB commit.* Classic dual-write problem: an Order is written to the database, message broker fails, downstream services never get notified.
- *Simutaneously call each service inside the order transaction.* Adds latency, everyone waits until each services finishes it's job. One service slows down the whole checkout process.
