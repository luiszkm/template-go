-- name: LockRole :one
SELECT id, name FROM roles WHERE id = $1 FOR UPDATE;

-- name: RolePermissions :many
SELECT permission FROM role_permissions WHERE role_id = $1 ORDER BY permission;

-- name: CountHolders :one
SELECT count(*)::int FROM user_roles WHERE role_id = $1;

-- name: RenameRole :exec
UPDATE roles SET name = $2 WHERE id = $1;

-- name: DeletePermissions :exec
DELETE FROM role_permissions WHERE role_id = $1;

-- name: InsertPermissions :exec
INSERT INTO role_permissions (role_id, permission) SELECT @role_id::uuid, unnest(@permissions::text[]);
