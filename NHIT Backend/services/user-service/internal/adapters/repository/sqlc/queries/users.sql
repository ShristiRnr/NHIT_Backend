-- name: CreateUser :one
INSERT INTO users (
    tenant_id, name, email, password, department_id, designation_id,
    account_holder_name, bank_name, bank_account_number, ifsc_code, signature_url, is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: CreateUserWithVerification :one
INSERT INTO users (
    tenant_id, name, email, password, email_verified_at, department_id, designation_id,
    account_holder_name, bank_name, bank_account_number, ifsc_code, signature_url, is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByEmailAndTenant :one
SELECT * FROM users
WHERE tenant_id = $1 AND email = $2;

-- name: UpdateUser :one
UPDATE users
SET 
    name = $2, 
    email = $3, 
    password = $4, 
    department_id = $5,
    designation_id = $6,
    account_holder_name = $7,
    bank_name = $8,
    bank_account_number = $9,
    ifsc_code = $10,
    signature_url = $11,
    is_active = $12,
    deactivated_at = $13,
    deactivated_by = $14,
    deactivated_by_name = $15,
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;
-- name: UpdateUserProfile :one
UPDATE users
SET name = $2, email = $3, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password = $2, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserEmailVerification :one
UPDATE users
SET email_verified_at = $2, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserLastLogin :one
UPDATE users
SET last_login_at = $2, last_login_ip = $3, user_agent = $4, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserLastLogout :one
UPDATE users
SET last_logout_at = $2, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserDepartment :one
UPDATE users
SET department_id = $2, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserDesignation :one
UPDATE users
SET designation_id = $2, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = $1;

-- name: ListUsersByTenant :many
SELECT * FROM users
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsersByTenant :one
SELECT COUNT(*) FROM users
WHERE tenant_id = $1;

-- name: CountAllUsers :one
SELECT COUNT(*) FROM users;

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1) AS exists;

-- name: UserExistsByID :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1) AS exists;

-- name: ListUsersByDepartment :many
SELECT * FROM users
WHERE department_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUsersByDesignation :many
SELECT * FROM users
WHERE designation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SearchUsersByName :many
SELECT * FROM users
WHERE LOWER(name) LIKE LOWER('%' || $1 || '%')
ORDER BY name ASC
LIMIT $2 OFFSET $3;

-- name: SearchUsersByEmail :many
SELECT * FROM users
WHERE LOWER(email) LIKE LOWER('%' || $1 || '%')
ORDER BY email ASC
LIMIT $2 OFFSET $3;
