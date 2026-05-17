-- name: GetProductByID :one
SELECT id, name, price_amount, price_currency,
       inventory_total, inventory_available,
       created_at, updated_at
  FROM products
 WHERE id = $1;

-- name: ListProductsByCursor :many
-- Cursor pagination per ADR 006. The cursor is (created_at, id) so we can
-- resolve ties deterministically. First page passes the sentinel
-- (now() + interval '1 day', uuid_nil()) so all rows match.
SELECT id, name, price_amount, price_currency,
       inventory_total, inventory_available,
       created_at, updated_at
  FROM products
 WHERE (created_at, id) < ($1::timestamptz, $2::uuid)
 ORDER BY created_at DESC, id DESC
 LIMIT $3;

-- name: InsertProduct :one
INSERT INTO products (
    id, name, price_amount, price_currency,
    inventory_total, inventory_available
) VALUES (
    $1, $2, $3, $4, $5, $5
)
RETURNING id, name, price_amount, price_currency,
          inventory_total, inventory_available,
          created_at, updated_at;
