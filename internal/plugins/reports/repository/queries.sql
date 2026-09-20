-- name: InsertReport :one
INSERT INTO community_reports (id, reporter_id, target, category, message)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListReportsByReporter :many
SELECT * FROM community_reports
WHERE reporter_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListReports :many
-- All reports, newest first, optionally filtered by status ('' means every status).
SELECT * FROM community_reports
WHERE ($3::text = '' OR status = $3)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetReport :one
SELECT * FROM community_reports WHERE id = $1;

-- name: CloseReport :one
UPDATE community_reports
SET status = 'closed', updated_at = now()
WHERE id = $1
RETURNING *;
