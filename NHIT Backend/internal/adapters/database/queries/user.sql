-- name: CreateUserWithVerification :one
INSERT INTO users (tenant_id, name, email, password, email_verified_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE user_id = $1;

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
