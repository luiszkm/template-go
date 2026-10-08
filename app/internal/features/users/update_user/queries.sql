-- name: LockUser :one
SELECT id, email, name, created_at, deactivated_at FROM users WHERE id = $1 FOR UPDATE;

-- name: UpdateUser :one
UPDATE users SET email = $2, name = $3 WHERE id = $1
RETURNING id, email, name, created_at, deactivated_at;
