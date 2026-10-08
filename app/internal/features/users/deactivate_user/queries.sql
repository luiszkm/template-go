-- name: LockUser :one
SELECT id, email, name, deactivated_at FROM users WHERE id = $1 FOR UPDATE;

-- name: Deactivate :one
UPDATE users SET deactivated_at = now() WHERE id = $1 RETURNING deactivated_at;

-- name: DeleteUserSessions :execrows
DELETE FROM sessions WHERE user_id = $1;
