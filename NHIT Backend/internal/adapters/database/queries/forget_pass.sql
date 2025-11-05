-- Create password reset token
INSERT INTO password_resets (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING token;
