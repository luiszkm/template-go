-- name: PasswordHash :one
SELECT password_hash FROM users WHERE id = $1;

-- name: UpdatePasswordHash :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: DeleteOtherSessions :execrows
DELETE FROM sessions WHERE user_id = $1 AND token_hash <> $2;
