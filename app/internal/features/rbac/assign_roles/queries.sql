-- name: LockUser :one
SELECT id FROM users WHERE id = $1 FOR UPDATE;

-- name: RolesByID :many
SELECT id, name FROM roles WHERE id = ANY(@ids::uuid[]);

-- name: UserRoles :many
SELECT r.id, r.name FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = $1;

-- name: LockAdminRole :one
SELECT id FROM roles WHERE name = 'admin' FOR UPDATE;

-- name: CountOtherActiveHolders :one
SELECT count(*)::int FROM user_roles ur JOIN users u ON u.id = ur.user_id
WHERE ur.role_id = @role_id AND u.deactivated_at IS NULL AND u.id <> @user_id;

-- name: ClearUserRoles :exec
DELETE FROM user_roles WHERE user_id = $1;

-- name: InsertUserRoles :exec
INSERT INTO user_roles (user_id, role_id) SELECT @user_id::uuid, unnest(@role_ids::uuid[]);

-- name: DeleteUserSessions :execrows
DELETE FROM sessions WHERE user_id = $1;
