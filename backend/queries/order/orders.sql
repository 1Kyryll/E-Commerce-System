-- name: InsertOrder :one
INSERT INTO orders (
    id, idempotency_key, user_id,
    total_amount, total_currency, status
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, idempotency_key, user_id,
          total_amount, total_currency, status,
          created_at, updated_at;

-- name: GetOrderByID :one
SELECT id, idempotency_key, user_id,
       total_amount, total_currency, status,
       created_at, updated_at
  FROM orders
 WHERE id = $1;

-- name: GetOrderByIdempotencyKey :one
SELECT id, idempotency_key, user_id,
       total_amount, total_currency, status,
       created_at, updated_at
  FROM orders
 WHERE idempotency_key = $1;

-- name: InsertOrderItem :one
INSERT INTO order_items (
    id, order_id, product_id, reservation_id,
    quantity, unit_price_amount, unit_price_currency
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, order_id, product_id, reservation_id,
          quantity, unit_price_amount, unit_price_currency;

-- name: ListOrderItems :many
SELECT id, order_id, product_id, reservation_id,
       quantity, unit_price_amount, unit_price_currency
  FROM order_items
 WHERE order_id = $1;

-- name: GetProductPriceForOrder :one
-- Used by PlaceOrder to compute order total. Order service legitimately
-- reads from products (see docs/system-design.md: ORD --> PDB).
SELECT id, price_amount, price_currency
  FROM products
 WHERE id = $1;
