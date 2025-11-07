-- name: CreateOrganization :one
INSERT INTO organizations (
    org_id, tenant_id, name, code, database_name,
    description, logo, is_active, created_by,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations
WHERE org_id = $1 AND tenant_id = $2;

-- name: GetOrganizationByCode :one
SELECT * FROM organizations
WHERE code = $1 AND tenant_id = $2;

-- name: UpdateOrganization :one
UPDATE organizations
SET 
    name = CASE WHEN @name::text IS NOT NULL THEN @name ELSE name END,
    description = CASE WHEN @description::text IS NOT NULL THEN @description ELSE description END,
    logo = CASE WHEN @logo::text IS NOT NULL THEN @logo ELSE logo END,
    is_active = CASE WHEN @is_active::boolean IS NOT NULL THEN @is_active ELSE is_active END,
    updated_at = @updated_at
WHERE org_id = @org_id AND tenant_id = @tenant_id
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM organizations
WHERE org_id = $1 AND tenant_id = $2;

-- name: ListOrganizationsByTenant :many
SELECT * FROM organizations
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountOrganizationsByTenant :one
SELECT COUNT(*) FROM organizations
WHERE tenant_id = $1;

-- name: CheckOrganizationCodeExists :one
SELECT EXISTS(
    SELECT 1 FROM organizations
    WHERE code = $1 AND tenant_id = $2
) AS exists;

-- name: ToggleOrganizationStatus :one
UPDATE organizations
SET is_active = NOT is_active,
    updated_at = CURRENT_TIMESTAMP
WHERE org_id = $1 AND tenant_id = $2
RETURNING *;
