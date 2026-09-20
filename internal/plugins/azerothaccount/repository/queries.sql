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

-- name: UpsertAccountClaim :one
INSERT INTO account_claims (user_id, account_username, code_hash, expires_at, attempts, updated_at)
VALUES ($1, $2, $3, $4, 0, now())
ON CONFLICT (user_id) DO UPDATE
SET account_username = EXCLUDED.account_username,
    code_hash = EXCLUDED.code_hash,
    expires_at = EXCLUDED.expires_at,
    attempts = 0,
    updated_at = now()
RETURNING *;

-- name: GetAccountClaim :one
SELECT * FROM account_claims WHERE user_id = $1;

-- name: IncrementAccountClaimAttempts :one
UPDATE account_claims
SET attempts = attempts + 1, updated_at = now()
WHERE user_id = $1
RETURNING *;

-- name: DeleteAccountClaim :exec
DELETE FROM account_claims WHERE user_id = $1;

-- name: ListAccountClaims :many
SELECT * FROM account_claims ORDER BY created_at DESC LIMIT $1 OFFSET $2;
