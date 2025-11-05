-- name: CreateRole :one
INSERT INTO roles (tenant_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetRole :one
SELECT *
FROM roles
WHERE role_id = $1;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, updated_at = now()
WHERE role_id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = $1;

-- name: AssignPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: ListRolesByTenant :many
SELECT *
FROM roles
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListRolesOfUser :many
SELECT r.*
FROM roles r
JOIN user_roles ur ON ur.role_id = r.role_id
WHERE ur.user_id = $1;

-- name: ListPermissionsOfUserViaRoles :many
SELECT p.*
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.permission_id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1;
