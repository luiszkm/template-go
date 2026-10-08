-- name: GetRole :one
SELECT r.id, r.name,
    COALESCE(array_agg(rp.permission ORDER BY rp.permission) FILTER (WHERE rp.permission IS NOT NULL), '{}')::text[] AS permissions,
    (SELECT count(*) FROM user_roles ur WHERE ur.role_id = r.id)::int AS user_count
FROM roles r
LEFT JOIN role_permissions rp ON rp.role_id = r.id
WHERE r.id = $1
GROUP BY r.id;
