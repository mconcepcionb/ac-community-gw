-- Seeds an internal role with every permission the gateway exposes and maps a
-- Discord role to it, so the test frontend can run the AzerothCore commands
-- without a 403.
--
-- Usage (host psql):
--   psql "$ACGW_DATABASE_URL" -v discord_role_id=<role id> -f scripts/seed_demo_admin.sql
--
-- Usage (docker compose postgres):
--   Get-Content scripts/seed_demo_admin.sql -Raw |
--     docker compose exec -T postgres psql -U acgw -d acgw \
--       -v discord_role_id=<role id> -v internal_role=ac-core.admin
--
-- Variables:
--   discord_role_id  (required) Discord role id, copied with Developer Mode
--                    enabled (right-click the role -> Copy Role ID).
--   internal_role    (optional) internal role name, defaults to ac-core.admin.
--   include_admin    (optional) set to "true" to also grant azeroth.admin.* and
--                    store.admin.* scopes. Defaults to "false" so the demo role
--                    cannot administer the server.
--
-- The gateway loads role grants at startup, so restart it and log in again
-- afterwards.

\if :{?discord_role_id}
\else
\echo 'ERROR: pass -v discord_role_id=<discord role id>'
\quit
\endif

\if :{?internal_role}
\else
\set internal_role 'ac-core.admin'
\endif

\if :{?include_admin}
\else
\set include_admin 'false'
\endif

BEGIN;

INSERT INTO permissions (name, description, owner) VALUES
    ('identity.self.read', 'Read the authenticated user''s own profile', 'identity-discord'),
    ('identity.session.revoke', 'Revoke user sessions', 'identity-discord'),
    ('identity.user.list', 'Search community users', 'identity-discord'),
    ('azeroth.info.public.read', 'Read public AzerothCore server information', 'azeroth-info'),
    ('azeroth.info.private.read', 'Read private AzerothCore server information', 'azeroth-info'),
    ('azeroth.account.read', 'Read linked AzerothCore account information', 'azeroth-account'),
    ('azeroth.account.manage', 'Create and manage AzerothCore accounts', 'azeroth-account'),
    ('azeroth.account.list', 'List AzerothCore login accounts', 'azeroth-account'),
    ('azeroth.account.link', 'Create and remove community user / AzerothCore account links', 'azeroth-account'),
    ('azeroth.character.list', 'Read AzerothCore characters', 'azeroth-character'),
    ('azeroth.mail.send', 'Send in-game mail, items and money', 'azeroth-character'),
    ('azeroth.item.list', 'Search the AzerothCore item catalog', 'azeroth-item'),
    ('azeroth.admin.accounts.read', 'Read AzerothCore accounts as an administrator', 'azeroth-admin'),
    ('azeroth.admin.accounts.ban', 'Ban and unban AzerothCore accounts', 'azeroth-admin'),
    ('azeroth.admin.accounts.gmlevel', 'Change AzerothCore GM level', 'azeroth-admin'),
    ('azeroth.admin.players.read', 'List online players', 'azeroth-admin'),
    ('azeroth.admin.players.kick', 'Kick online players', 'azeroth-admin'),
    ('azeroth.admin.players.mute', 'Mute and unmute players', 'azeroth-admin'),
    ('azeroth.admin.characters.ban', 'Ban and unban characters', 'azeroth-admin'),
    ('azeroth.admin.announce', 'Broadcast announcements', 'azeroth-admin'),
    ('store.catalog.read', 'Read the store product catalog', 'azeroth-store'),
    ('store.wallet.read', 'Read the authenticated user''s wallet balance', 'azeroth-store'),
    ('store.orders.read', 'Read the authenticated user''s orders', 'azeroth-store'),
    ('store.purchase', 'Buy store products', 'azeroth-store'),
    ('store.admin.wallets', 'Grant points to any wallet', 'azeroth-store'),
    ('store.admin.products', 'Manage the store product catalog', 'azeroth-store')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (name, description) VALUES
    (:'internal_role', 'Demo administrator')
ON CONFLICT (name) DO NOTHING;

-- Grant the role the registered permissions. Administrative scopes are only
-- granted when include_admin=true.
INSERT INTO role_permissions (role, permission)
SELECT :'internal_role', name FROM permissions
WHERE :'include_admin' = 'true'
   OR (name NOT LIKE 'azeroth.admin.%' AND name NOT LIKE 'store.admin.%')
ON CONFLICT DO NOTHING;

-- Optional database mapping; equivalent static mappings can be provided with
-- ACGW_DISCORD_ROLE_MAPPINGS=<discord_role_id>=<internal_role>.
INSERT INTO discord_role_mappings (discord_role_id, role)
VALUES (:'discord_role_id', :'internal_role')
ON CONFLICT (discord_role_id) DO UPDATE
SET role = EXCLUDED.role,
    updated_at = now();

COMMIT;
