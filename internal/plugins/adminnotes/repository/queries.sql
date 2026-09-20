-- name: InsertAnnotation :one
INSERT INTO admin_annotations (id, target_type, target_id, author_id, body)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAnnotation :one
SELECT * FROM admin_annotations WHERE id = $1;

-- name: ListAnnotations :many
SELECT * FROM admin_annotations
WHERE target_type = $1 AND target_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateAnnotation :one
UPDATE admin_annotations
SET body = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteAnnotation :execrows
DELETE FROM admin_annotations WHERE id = $1;
