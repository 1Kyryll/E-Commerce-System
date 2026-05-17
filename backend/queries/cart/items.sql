-- name: GetCartByUserID :one
SELECT id, user_id, created_at, updated_at
  FROM carts
 WHERE user_id = $1;

-- name: UpsertCart :one
INSERT INTO carts (id, user_id)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE
   SET updated_at = now()
RETURNING id, user_id, created_at, updated_at;

-- name: ListCartItems :many
SELECT id, cart_id, product_id, quantity, added_at
  FROM cart_items
 WHERE cart_id = $1
 ORDER BY added_at ASC;

-- name: UpsertCartItem :one
INSERT INTO cart_items (id, cart_id, product_id, quantity)
VALUES ($1, $2, $3, $4)
ON CONFLICT (cart_id, product_id) DO UPDATE
   SET quantity = cart_items.quantity + EXCLUDED.quantity
RETURNING id, cart_id, product_id, quantity, added_at;

-- name: RemoveCartItem :exec
DELETE FROM cart_items
 WHERE cart_id = $1 AND product_id = $2;

-- name: ClearCart :exec
DELETE FROM cart_items WHERE cart_id = $1;
