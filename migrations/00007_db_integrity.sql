-- +goose Up

-- Account usernames are case-insensitive in AzerothCore: enforce it with a
-- unique index on the lower-cased value instead of a case-sensitive UNIQUE.
ALTER TABLE azeroth_account_links
    DROP CONSTRAINT IF EXISTS azeroth_account_links_account_username_key;
CREATE UNIQUE INDEX azeroth_account_links_username_lower_key
    ON azeroth_account_links (lower(account_username));

-- Keep the account id type consistent with the store tables (upstream
-- account.id is an unsigned int).
ALTER TABLE azeroth_account_links
    ALTER COLUMN account_id TYPE bigint;

-- Financial history must not be destroyed by deleting a community user.
ALTER TABLE store_wallets DROP CONSTRAINT IF EXISTS store_wallets_user_id_fkey;
ALTER TABLE store_wallets ADD CONSTRAINT store_wallets_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES community_users (id) ON DELETE RESTRICT;

ALTER TABLE store_wallet_entries DROP CONSTRAINT IF EXISTS store_wallet_entries_user_id_fkey;
ALTER TABLE store_wallet_entries ADD CONSTRAINT store_wallet_entries_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES community_users (id) ON DELETE RESTRICT;

ALTER TABLE store_orders DROP CONSTRAINT IF EXISTS store_orders_user_id_fkey;
ALTER TABLE store_orders ADD CONSTRAINT store_orders_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES community_users (id) ON DELETE RESTRICT;

-- Wallet entries reference the order that produced them.
ALTER TABLE store_wallet_entries
    ADD CONSTRAINT store_wallet_entries_order_id_fkey
    FOREIGN KEY (order_id) REFERENCES store_orders (id) ON DELETE SET NULL;
CREATE INDEX store_wallet_entries_order_id_idx ON store_wallet_entries (order_id);

-- Uniqueness of a Discord id lives on community_users; the identity row is
-- keyed by user.
ALTER TABLE discord_identities
    DROP CONSTRAINT IF EXISTS discord_identities_discord_id_key;
ALTER TABLE discord_identities
    ADD CONSTRAINT discord_identities_user_discord_key UNIQUE (user_id, discord_id);

-- +goose Down

ALTER TABLE discord_identities
    DROP CONSTRAINT IF EXISTS discord_identities_user_discord_key;
ALTER TABLE discord_identities
    ADD CONSTRAINT discord_identities_discord_id_key UNIQUE (discord_id);

DROP INDEX IF EXISTS store_wallet_entries_order_id_idx;
ALTER TABLE store_wallet_entries
    DROP CONSTRAINT IF EXISTS store_wallet_entries_order_id_fkey;

ALTER TABLE store_orders DROP CONSTRAINT IF EXISTS store_orders_user_id_fkey;
ALTER TABLE store_orders ADD CONSTRAINT store_orders_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES community_users (id) ON DELETE CASCADE;

ALTER TABLE store_wallet_entries DROP CONSTRAINT IF EXISTS store_wallet_entries_user_id_fkey;
ALTER TABLE store_wallet_entries ADD CONSTRAINT store_wallet_entries_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES community_users (id) ON DELETE CASCADE;

ALTER TABLE store_wallets DROP CONSTRAINT IF EXISTS store_wallets_user_id_fkey;
ALTER TABLE store_wallets ADD CONSTRAINT store_wallets_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES community_users (id) ON DELETE CASCADE;

ALTER TABLE azeroth_account_links
    ALTER COLUMN account_id TYPE integer;

DROP INDEX IF EXISTS azeroth_account_links_username_lower_key;
ALTER TABLE azeroth_account_links
    ADD CONSTRAINT azeroth_account_links_account_username_key UNIQUE (account_username);
