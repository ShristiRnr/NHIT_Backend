INSERT INTO users (tenant_id, name, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, name, email;
