
-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING token, user_id, expires_at, created_at;


-- name: GetUserIDByRefreshToken :one
SELECT user_id
FROM refresh_tokens
WHERE token = $1;


-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;