-- Delete refresh token
DELETE FROM refresh_tokens
WHERE token = $1;
