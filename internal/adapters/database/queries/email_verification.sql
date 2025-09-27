-- name: InsertEmailVerification :one
INSERT INTO email_verifications (user_id, token, created_at)
VALUES ($1, $2, NOW())
RETURNING *;

-- name: GetEmailVerificationByToken :one
SELECT * FROM email_verifications WHERE token = $1;

-- name: DeleteEmailVerificationByUserID :exec
DELETE FROM email_verifications WHERE user_id = $1;
