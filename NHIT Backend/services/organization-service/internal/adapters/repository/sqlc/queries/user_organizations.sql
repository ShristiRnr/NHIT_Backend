-- name: CreateUserOrganization :one
INSERT INTO user_organizations (
    user_id, org_id, role_id, is_current_context,
    joined_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserOrganizations :many
SELECT * FROM user_organizations
WHERE user_id = $1
ORDER BY joined_at DESC;

-- name: GetOrganizationUsers :many
SELECT * FROM user_organizations
WHERE org_id = $1
ORDER BY joined_at DESC;

-- name: GetUserOrganization :one
SELECT * FROM user_organizations
WHERE user_id = $1 AND org_id = $2;

-- name: UpdateUserOrganizationRole :one
UPDATE user_organizations
SET role_id = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $2 AND org_id = $3
RETURNING *;

-- name: SetCurrentContext :exec
UPDATE user_organizations
SET is_current_context = CASE 
    WHEN org_id = $2 THEN true
    ELSE false
END
WHERE user_id = $1;

-- name: DeleteUserOrganization :exec
DELETE FROM user_organizations
WHERE user_id = $1 AND org_id = $2;

-- name: CountUserOrganizations :one
SELECT COUNT(*) FROM user_organizations
WHERE user_id = $1;
