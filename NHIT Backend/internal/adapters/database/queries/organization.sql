-- name: CreateOrganization :one
INSERT INTO organizations (tenant_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrganization :one
SELECT *
FROM organizations
WHERE org_id = $1;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2,
    updated_at = now()
WHERE org_id = $1
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM organizations
WHERE org_id = $1;

-- name: ListOrganizationsByTenant :many
SELECT *
FROM organizations
WHERE tenant_id = $1;
