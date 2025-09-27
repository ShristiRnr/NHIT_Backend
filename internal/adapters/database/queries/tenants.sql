-- name: CreateTenant :one
INSERT INTO tenants (name, super_admin_user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetTenant :one
SELECT *
FROM tenants
WHERE tenant_id = $1;
