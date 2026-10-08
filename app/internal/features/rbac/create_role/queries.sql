-- name: InsertRole :one
INSERT INTO roles (name) VALUES ($1) RETURNING id, name;

-- name: InsertPermissions :exec
INSERT INTO role_permissions (role_id, permission) SELECT @role_id::uuid, unnest(@permissions::text[]);
