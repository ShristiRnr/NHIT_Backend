-- name: PaginatedUsersByTenant :many
SELECT *
FROM users
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUsersByTenant :one
SELECT COUNT(*)
FROM users
WHERE tenant_id = $1;
