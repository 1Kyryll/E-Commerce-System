### 4. Cart: persistent, not session-bound

According to [System Requirements](../system-requirements.md.md), `Cart` must be persistent. Therefore, every add-to-cart request is saved to the `cart` table keyed by `User ID`. 

**Tradeoff:** Cart writes become DB writes (more load), but cart-recovery is free — a user who closes the tab and comes back tomorrow finds their cart. 

**Alternatives:**
- *Browser cookie / localStorage.* User loses their cart on device change or browser clear.
- *Server-side session in Redis.* Lost on Redis eviction or restart.
