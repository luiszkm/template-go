-- name: LockUser :one
SELECT id, email, name, deactivated_at FROM users WHERE id = $1 FOR UPDATE;

-- name: Activate :exec
UPDATE users SET deactivated_at = NULL WHERE id = $1;
