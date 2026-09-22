-- Grant demo points to a community user, identified by Discord id.
--
-- Usage:
--   psql "$ACGW_DATABASE_URL" -v discord_id=<id> -v points=1000 \
--     -f scripts/seed_demo_grant.sql
--
-- The user must exist (log in once first). Idempotent enough for a demo: it
-- adds the points to the current balance.

\if :{?discord_id}
\else
\echo 'ERROR: pass -v discord_id=<discord user id>'
\quit
\endif

\if :{?points}
\else
\set points 1000
\endif

WITH target AS (
    SELECT id FROM community_users WHERE discord_id = :'discord_id'
),
upsert AS (
    INSERT INTO store_wallets (user_id, balance)
    SELECT id, :'points'::bigint FROM target
    ON CONFLICT (user_id) DO UPDATE
        SET balance = store_wallets.balance + EXCLUDED.balance,
            updated_at = now()
    RETURNING user_id, balance
)
INSERT INTO store_wallet_entries (user_id, delta, balance_after, reason)
SELECT user_id, :'points'::bigint, balance, 'demo welcome grant' FROM upsert;

\echo 'demo: wallet funded (or skipped when the user does not exist yet)'
