-- name: CreateDesignation :one
INSERT INTO designations (id, name, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetDesignationByID :one
SELECT * FROM designations WHERE id = $1;

-- name: UpdateDesignation :one
UPDATE designations
SET name = $2, description = $3, updated_at = $4
WHERE id = $1
RETURNING *;

-- name: DeleteDesignation :exec
DELETE FROM designations WHERE id = $1;

-- name: ListDesignations :many
SELECT id, name, description, created_at, updated_at
FROM designations
ORDER BY name ASC
LIMIT $1 OFFSET $2;
