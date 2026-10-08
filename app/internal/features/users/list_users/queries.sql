-- name: ListUsers :many
SELECT id, email, name, created_at, deactivated_at FROM users ORDER BY email LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM users;
