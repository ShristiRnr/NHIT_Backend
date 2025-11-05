-- Check email and password hash
SELECT id, name, email FROM users
WHERE email = $1;

-- Update last login
UPDATE users
SET last_login_at = NOW(), last_login_ip = $2
WHERE id = $1;
