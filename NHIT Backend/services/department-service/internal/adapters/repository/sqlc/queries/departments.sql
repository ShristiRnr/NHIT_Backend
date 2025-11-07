-- name: CreateDepartment :one
INSERT INTO departments (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetDepartmentByID :one
SELECT * FROM departments WHERE id = $1;

-- name: GetDepartmentByName :one
SELECT * FROM departments WHERE name = $1;

-- name: UpdateDepartment :one
UPDATE departments
SET name = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDepartment :exec
DELETE FROM departments WHERE id = $1;

-- name: ListDepartments :many
SELECT * FROM departments 
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2;

-- name: CountDepartments :one
SELECT COUNT(*) FROM departments;

-- name: DepartmentExists :one
SELECT EXISTS(SELECT 1 FROM departments WHERE name = $1) AS exists;

-- name: DepartmentExistsByID :one
SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1) AS exists;
