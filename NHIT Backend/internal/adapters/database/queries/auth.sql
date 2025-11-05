-- Registration
-- name: CreateUser :one
INSERT INTO users (tenant_id, name, email, password)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmailAndTenant :one
SELECT *
FROM users
WHERE tenant_id = $1 AND email = $2;

-- Optional org-scoped fetch for login flows where org must be enforced
-- name: GetUserByEmailTenantAndOrg :one
SELECT u.*
FROM users u
JOIN user_organizations uo ON uo.user_id = u.user_id
WHERE u.tenant_id = $1 AND uo.org_id = $2 AND u.email = $3;

-- Email Verification
-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (user_id, expires_at)
VALUES ($1, $2)
RETURNING *;

-- name: CheckEmailVerificationToken :one
SELECT *
FROM email_verification_tokens
WHERE user_id = $1 AND token = $2 AND expires_at > NOW();

-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified_at = NOW(), updated_at = NOW()
WHERE user_id = $1;

-- name: DeleteEmailVerificationTokensByUser :exec
DELETE FROM email_verification_tokens
WHERE user_id = $1;

-- Password Reset
-- name: CreatePasswordResetToken :one
INSERT INTO password_resets (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING token;

-- name: GetPasswordResetToken :one
SELECT *
FROM password_resets
WHERE token = $1 AND expires_at > NOW();

-- name: DeletePasswordResetToken :exec
DELETE FROM password_resets
WHERE token = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = $2, updated_at = NOW()
WHERE user_id = $1;

-- Login / Logout helpers
-- name: RecordUserLogin :one
INSERT INTO user_login_history (user_id, ip_address, user_agent)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(), last_login_ip = $2, user_agent = $3, updated_at = NOW()
WHERE user_id = $1;

-- Refresh Tokens
-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: GetUserIDByRefreshToken :one
SELECT user_id
FROM refresh_tokens
WHERE token = $1 AND expires_at > NOW();

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;

-- Roles & Permissions for response enrichment
-- name: ListUserRoleNames :many
SELECT r.name
FROM roles r
JOIN user_roles ur ON ur.role_id = r.role_id
WHERE ur.user_id = $1
ORDER BY r.name;

-- name: ListUserPermissionNames :many
SELECT p.name
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.permission_id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1
ORDER BY p.name;

-- Org membership check
-- name: IsUserInOrganization :one
SELECT EXISTS (
  SELECT 1
  FROM user_organizations
  WHERE user_id = $1 AND org_id = $2
) AS exists;

-- Role assignment helpers for registration/updates
-- name: ListRoleIDsByNamesForTenant :many
SELECT role_id
FROM roles
WHERE tenant_id = $1 AND name = ANY($2::text[]);

-- name: DeleteRolesForUser :exec
DELETE FROM user_roles
WHERE user_id = $1;

-- name: AssignRolesToUserBulk :exec
INSERT INTO user_roles (user_id, role_id)
SELECT $1 AS user_id, unnest($2::uuid[]) AS role_id
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: UpdateUserLastLogout :exec
UPDATE users
SET last_logout_at = NOW(), updated_at = NOW()
WHERE user_id = $1;
