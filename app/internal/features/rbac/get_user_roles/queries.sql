-- name: UserExists :one
SELECT EXISTS (SELECT 1 FROM users WHERE id = $1);

-- name: UserRoles :many
SELECT r.id, r.name FROM user_roles ur JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = $1
ORDER BY lower(r.name), r.name;
