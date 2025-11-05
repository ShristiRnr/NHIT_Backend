-- name: CreatePasswordResetToken :one
INSERT INTO password_resets (email, token, expires_at, created_at)
VALUES ($1, $2, $3, NOW())
RETURNING token, email, expires_at, created_at;

-- name: GetPasswordResetByToken :one
SELECT token, email, expires_at, created_at
FROM password_resets
WHERE token = $1;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_resets
WHERE token = $1;
