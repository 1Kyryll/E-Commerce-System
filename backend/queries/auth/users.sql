-- name: InsertUser :one
INSERT INTO users (id, name, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, name, email, password_hash, created_at;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, created_at
  FROM users
 WHERE email = $1;

-- name: GetUserByID :one
SELECT id, name, email, password_hash, created_at
  FROM users
 WHERE id = $1;
