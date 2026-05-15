### 2. Reservation table for the payment window

We need a Reservations table in order to provide explicit and queryable reservations lifecycle. We rely on the reservations and not the absence of the inventory. Each reservation comes with a *15 minute TTL*.

**Tradeoff:** Clean-up worker for row expirations and an additional table. 

**Alternatives:** 
- *Decrement on payment success.* Fails on 10K users trying to purchase and 10 units of a product available at the moment. Both users see the item is available, after the payment succeeded the product is suddenly out of stock.
- *Decrement on add-to-cart.* Not every product added to cart will be purchased.
- *No Reservations, rely on Order status.* Forces frequent scans on the `orders` table which is exprensive, also comes up with complicated code.
