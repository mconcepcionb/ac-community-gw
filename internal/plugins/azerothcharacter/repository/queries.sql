-- name: UpsertVisibility :one
INSERT INTO character_visibility (character_name, user_id, public, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (character_name)
DO UPDATE SET public = EXCLUDED.public, user_id = EXCLUDED.user_id, updated_at = now()
RETURNING *;

-- name: GetVisibility :one
SELECT * FROM character_visibility WHERE character_name = $1;

-- name: ListVisibilityByUser :many
SELECT character_name, public FROM character_visibility WHERE user_id = $1;

-- name: ListPublicCharacterNames :many
SELECT character_name FROM character_visibility WHERE public = true;
