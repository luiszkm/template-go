-- name: RetryAfter :one
SELECT ceil(extract(epoch FROM (at + interval '15 minutes' - now())))::int AS seconds
FROM login_attempts
WHERE kind = sqlc.arg(kind) AND key = sqlc.arg(key) AND at > now() - interval '15 minutes'
ORDER BY at DESC
OFFSET sqlc.arg(limit_minus_one)::int
LIMIT 1;

-- name: UserByEmail :one
SELECT id, password_hash, deactivated_at FROM users WHERE email = $1;

-- name: RecordAttempt :exec
INSERT INTO login_attempts (kind, key) VALUES ($1, $2);

-- name: PruneAttempts :exec
DELETE FROM login_attempts WHERE at <= now() - interval '15 minutes';

-- name: ClearEmailAttempts :exec
DELETE FROM login_attempts WHERE kind = 'email' AND key = $1;

-- name: UpdatePasswordHash :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = $1;
