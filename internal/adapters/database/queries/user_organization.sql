-- name: AddUserToOrganization :exec
INSERT INTO user_organizations (user_id, org_id, role_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, org_id) DO UPDATE
SET role_id = EXCLUDED.role_id;

-- name: ListUsersByOrganization :many
SELECT u.user_id, u.name, u.email, ur.role_id, r.name AS role_name
FROM users u
JOIN user_organizations ur ON ur.user_id = u.user_id
JOIN roles r ON r.role_id = ur.role_id
WHERE ur.org_id = $1;
