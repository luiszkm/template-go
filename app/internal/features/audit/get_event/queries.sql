-- name: GetEvent :one
SELECT e.id, e.occurred_at, e.action, e.actor_id, u.email AS actor_email,
    e.resource_type, e.resource_id, COALESCE(host(e.ip), '')::text AS ip, e.request_id, e.before, e.after
FROM audit_events e
LEFT JOIN users u ON u.id = e.actor_id
WHERE e.id = $1;
