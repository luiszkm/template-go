-- name: ListEvents :many
SELECT e.id, e.occurred_at, e.action, e.actor_id, u.email AS actor_email,
    e.resource_type, e.resource_id, COALESCE(host(e.ip), '')::text AS ip, e.request_id
FROM audit_events e
LEFT JOIN users u ON u.id = e.actor_id
WHERE (sqlc.narg('before')::bigint IS NULL OR e.id < sqlc.narg('before')::bigint)
  AND (sqlc.narg('action')::text IS NULL OR e.action = sqlc.narg('action')::text)
  AND (sqlc.narg('actor_id')::uuid IS NULL OR e.actor_id = sqlc.narg('actor_id')::uuid)
  AND (sqlc.narg('resource_type')::text IS NULL OR e.resource_type = sqlc.narg('resource_type')::text)
  AND (sqlc.narg('resource_id')::text IS NULL OR e.resource_id = sqlc.narg('resource_id')::text)
  AND (sqlc.narg('from_at')::timestamptz IS NULL OR e.occurred_at >= sqlc.narg('from_at')::timestamptz)
  AND (sqlc.narg('to_at')::timestamptz IS NULL OR e.occurred_at < sqlc.narg('to_at')::timestamptz)
ORDER BY e.id DESC
LIMIT sqlc.arg('row_limit')::int;
