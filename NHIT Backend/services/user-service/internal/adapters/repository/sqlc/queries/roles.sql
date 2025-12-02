-- name: CreateRole :one
INSERT INTO roles (tenant_id, parent_org_id, name, description, permissions, is_system_role, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetRole :one
SELECT * FROM roles
WHERE role_id = $1;

-- name: ListRolesByTenant :many
SELECT * FROM roles
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListRolesByTenantAndOrg :many
SELECT * FROM roles
WHERE tenant_id = $1
  AND parent_org_id = $2
ORDER BY created_at DESC;

-- name: ListRolesByOrganizationIncludingSystem :many
SELECT * FROM roles
WHERE tenant_id = $1
  AND (parent_org_id = $2 OR (is_system_role = TRUE AND parent_org_id IS NULL))
ORDER BY created_at DESC;

-- name: UpdateRole :one
UPDATE roles
SET name = $2,
    description = $3,
    permissions = $4,
    updated_at = NOW()
WHERE role_id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = $1;

-- name: ListRolesByIDs :many
SELECT * FROM roles
WHERE role_id = ANY($1::uuid[])
ORDER BY created_at DESC;
