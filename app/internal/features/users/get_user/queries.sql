-- name: UserByID :one
SELECT id, email, name, created_at, deactivated_at FROM users WHERE id = $1;
