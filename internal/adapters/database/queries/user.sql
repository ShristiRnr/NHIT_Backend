-- name: CreateUser :one
INSERT INTO users (tenant_id, name, email, password, email_verified_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE user_id = $1;

-- name: MarkEmailVerified :exec
UPDATE users
SET email_verified_at = NOW()
WHERE user_id = $1;

-- name: UpdateUserPassword :one
UPDATE users
SET password = $2,
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: ListUsersByTenant :many
SELECT *
FROM users
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateUser :one
UPDATE users
SET name = $2,
    email = $3,
    password = $4,
    updated_at = now()
WHERE user_id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = $1;

-- name: GetUserPassword :one
SELECT password
FROM users
WHERE user_id = $1;

-- name: GetUserByEmail :one
SELECT user_id, tenant_id, name, emp_id, number, email, password, active, account_holder, bank_name, bank_account, ifsc_code, designation_id, department_id, email_verified_at, last_login_at, last_logout_at, last_login_ip, user_agent, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByToken :one
SELECT u.*
FROM users u
JOIN sessions s ON s.user_id = u.user_id
WHERE s.token = $1
  AND s.expires_at > NOW();

-- name: GetUserPermissions :many
SELECT p.name AS permission_name
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.permission_id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1;
