-- name: CreateDesignation :one
INSERT INTO designations (name, description)
VALUES ($1, $2)
RETURNING id, name, description, created_at, updated_at;

-- name: GetDesignation :one
SELECT id, name, description, created_at, updated_at
FROM designations
WHERE id = $1;

-- name: UpdateDesignation :one
UPDATE designations
SET name = $2,
    description = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteDesignation :exec
DELETE FROM designations
WHERE id = $1;

-- name: ListDesignations :many
SELECT id, name, description, created_at, updated_at
FROM designations
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
