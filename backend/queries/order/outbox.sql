-- name: InsertOutboxEvent :one
INSERT INTO outbox (
    id, aggregate_type, aggregate_id, event_type, payload
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, aggregate_type, aggregate_id, event_type, payload, created_at, published_at;

-- name: ListUnpublishedOutboxEvents :many
SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at
  FROM outbox
 WHERE published_at IS NULL
 ORDER BY created_at ASC
 LIMIT $1;

-- name: MarkOutboxEventPublished :exec
UPDATE outbox
   SET published_at = now()
 WHERE id = $1;
