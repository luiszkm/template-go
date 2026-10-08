-- name: UserByID :one
SELECT id, email, name FROM users WHERE id = $1;
