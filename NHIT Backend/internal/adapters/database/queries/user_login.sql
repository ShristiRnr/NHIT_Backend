-- name: CreateLoginHistory :one
INSERT INTO user_login_history (user_id, ip_address, user_agent)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListUserLoginHistories :many
SELECT *
FROM user_login_history
WHERE user_id = $1
ORDER BY login_time DESC
LIMIT $2 OFFSET $3;
