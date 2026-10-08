-- name: LockRole :one
SELECT id, name FROM roles WHERE id = $1 FOR UPDATE;

-- name: RolePermissions :many
SELECT permission FROM role_permissions WHERE role_id = $1 ORDER BY permission;

-- name: CountHolders :one
SELECT count(*)::int FROM user_roles WHERE role_id = $1;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;
