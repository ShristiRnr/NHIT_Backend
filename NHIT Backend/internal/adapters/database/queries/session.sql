-- name: CreateSession :one
INSERT INTO sessions (user_id, session_token, expires_at)
VALUES ($1, $2, $3)
RETURNING session_id, user_id, session_token, created_at, expires_at;

-- name: GetSession :one
SELECT session_id, user_id, session_token, created_at, expires_at
FROM sessions
WHERE session_id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE session_id = $1;
