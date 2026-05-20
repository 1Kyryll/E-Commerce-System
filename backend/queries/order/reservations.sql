-- name: DecrementInventoryAndCreateReservation :one
-- Atomic per ADR 009: only inserts the reservation if the inventory
-- decrement matched a row. Returns the new reservation, or no rows
-- when inventory was insufficient.
WITH decremented AS (
    UPDATE products
       SET inventory_available = inventory_available - $1,
           updated_at = now()
     WHERE id = $2
       AND inventory_available >= $1
    RETURNING id
)
INSERT INTO reservations (
    id, idempotency_key, product_id, user_id,
    quantity, status, expires_at
)
SELECT $3, $4, $2, $5, $1, 'active', now() + INTERVAL '15 minutes'
  FROM decremented
RETURNING id, idempotency_key, product_id, user_id, quantity,
          status, expires_at, created_at;

-- name: ConsumeReservationByIdempotencyKey :one
UPDATE reservations
   SET status = 'consumed', consumed_at = now()
 WHERE idempotency_key = $1
   AND status = 'active'
RETURNING id, product_id, user_id, quantity;

-- name: ReleaseReservationAndRestoreInventory :exec
-- Releases an active reservation and adds quantity back to inventory_available.
WITH released AS (
    UPDATE reservations
       SET status = 'released', released_at = now()
     WHERE reservations.id = $1
       AND reservations.status = 'active'
    RETURNING product_id, quantity
)
UPDATE products p
   SET inventory_available = p.inventory_available + r.quantity,
       updated_at = now()
  FROM released r
 WHERE p.id = r.product_id;

-- name: ListExpiredActiveReservations :many
SELECT id, product_id, user_id, quantity, expires_at
  FROM reservations
 WHERE status = 'active'
   AND expires_at < now()
 ORDER BY expires_at ASC
 LIMIT $1;

-- name: GetReservationByIdempotencyKey :one
SELECT id, idempotency_key, product_id, user_id, quantity,
       status, expires_at, created_at, consumed_at, released_at
  FROM reservations
 WHERE idempotency_key = $1;

-- name: CheckReservationPreconditions :one
-- Looks up an existing reservation by idempotency_key and checks whether
-- the product exists. The repo uses this BEFORE DecrementInventoryAndCreateReservation
-- to disambiguate the three failure modes that all otherwise collapse to
-- the atomic CTE returning pgx.ErrNoRows:
--   1) product missing       -> product_exists = false
--   2) idempotency replay    -> existing_reservation_id IS NOT NULL
--   3) insufficient inventory -> both fields above are "fine", and the
--                                 follow-up CTE returns ErrNoRows
SELECT
  EXISTS(SELECT 1 FROM products WHERE products.id = $1) AS product_exists,
  COALESCE(
    (SELECT reservations.id FROM reservations WHERE reservations.idempotency_key = $2),
    '00000000-0000-0000-0000-000000000000'::uuid
  )::uuid AS existing_reservation_id;
