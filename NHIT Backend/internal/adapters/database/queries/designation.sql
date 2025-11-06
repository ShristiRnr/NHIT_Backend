-- name: CreateDesignation :one
INSERT INTO designations (id, name, description, slug, is_active, parent_id, level, user_count, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetDesignationByID :one
SELECT * FROM designations WHERE id = $1;

-- name: GetDesignationBySlug :one
SELECT * FROM designations WHERE slug = $1;

-- name: GetDesignationByName :one
SELECT * FROM designations WHERE LOWER(name) = LOWER($1);

-- name: UpdateDesignation :one
UPDATE designations
SET name = $2, description = $3, slug = $4, is_active = $5, parent_id = $6, level = $7, updated_at = $8
WHERE id = $1
RETURNING *;

-- name: DeleteDesignation :exec
DELETE FROM designations WHERE id = $1;

-- name: ListDesignations :many
SELECT * FROM designations
WHERE 
    ($1::boolean = false OR is_active = true)
    AND ($2::uuid IS NULL OR parent_id = $2 OR ($2 = '00000000-0000-0000-0000-000000000000'::uuid AND parent_id IS NULL))
    AND ($3::text = '' OR LOWER(name) LIKE LOWER('%' || $3 || '%') OR LOWER(description) LIKE LOWER('%' || $3 || '%'))
ORDER BY level ASC, name ASC
LIMIT $4 OFFSET $5;

-- name: CountDesignations :one
SELECT COUNT(*) FROM designations
WHERE 
    ($1::boolean = false OR is_active = true)
    AND ($2::uuid IS NULL OR parent_id = $2 OR ($2 = '00000000-0000-0000-0000-000000000000'::uuid AND parent_id IS NULL))
    AND ($3::text = '' OR LOWER(name) LIKE LOWER('%' || $3 || '%') OR LOWER(description) LIKE LOWER('%' || $3 || '%'));

-- name: CheckDesignationExists :one
SELECT EXISTS(
    SELECT 1 FROM designations 
    WHERE LOWER(name) = LOWER($1) 
    AND ($2::uuid IS NULL OR id != $2)
);

-- name: CheckSlugExists :one
SELECT EXISTS(
    SELECT 1 FROM designations 
    WHERE slug = $1 
    AND ($2::uuid IS NULL OR id != $2)
);

-- name: GetDesignationChildren :many
SELECT * FROM designations WHERE parent_id = $1 ORDER BY name ASC;

-- name: GetDesignationUsersCount :one
SELECT COUNT(*) FROM users WHERE designation_id = $1;

-- name: UpdateDesignationUserCount :exec
UPDATE designations SET user_count = $2, updated_at = $3 WHERE id = $1;

-- name: GetDesignationLevel :one
SELECT level FROM designations WHERE id = $1;

-- name: CalculateDesignationLevel :one
SELECT COALESCE((SELECT level + 1 FROM designations WHERE id = $1), 0);

-- name: GetActiveDesignations :many
SELECT * FROM designations WHERE is_active = true ORDER BY level ASC, name ASC;

-- name: GetRootDesignations :many
SELECT * FROM designations WHERE parent_id IS NULL ORDER BY name ASC;

-- name: GetDesignationsByParent :many
SELECT * FROM designations WHERE parent_id = $1 ORDER BY name ASC;

-- name: DeactivateDesignation :exec
UPDATE designations SET is_active = false, updated_at = $2 WHERE id = $1;

-- name: ActivateDesignation :exec
UPDATE designations SET is_active = true, updated_at = $2 WHERE id = $1;
