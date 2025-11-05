-- name: CreateDepartment :one
INSERT INTO departments (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetDepartment :one
SELECT * FROM departments WHERE id = $1 LIMIT 1;

-- name: ListDepartments :many
SELECT * FROM departments ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateDepartment :one
UPDATE departments
SET name = $2, description = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDepartment :exec
DELETE FROM departments WHERE id = $1;
