### 3. Idempotency on order creation

**Decision.** Every order-creation query should be idempotent(multiple queries result same as one) and include `Idempotency-Key` header. The key is enforced as a unique constraint on the `reservations` table. A retried request with the same key returns the original order, not a new one.

**Tradeoff:** Client must generate and persist idempotency keys per checkout attempt.

**Alternatives:**
- *Trust the client.* Fails on every retry, network blip, or back-button.
- *Deduplicate by user + product + recent window.* Brittle — two intentional purchases of the same item within the window would be silently merged.
