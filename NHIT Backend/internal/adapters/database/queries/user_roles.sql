-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: ListRolesForUser :many
SELECT r.role_id, r.name
FROM roles r
JOIN user_roles ur ON r.role_id = ur.role_id
WHERE ur.user_id = $1;
