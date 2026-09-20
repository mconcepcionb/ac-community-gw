-- name: UpsertAccountLink :one
INSERT INTO azeroth_account_links (user_id, account_username, account_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO UPDATE
SET account_username = EXCLUDED.account_username,
    account_id = EXCLUDED.account_id,
    updated_at = now()
RETURNING *;

-- name: GetAccountLinkByUserID :one
SELECT * FROM azeroth_account_links WHERE user_id = $1;

-- name: ListAccountLinks :many
SELECT * FROM azeroth_account_links ORDER BY linked_at DESC;

-- name: DeleteAccountLink :exec
DELETE FROM azeroth_account_links WHERE user_id = $1;
