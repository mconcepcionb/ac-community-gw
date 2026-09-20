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
--                    gw.store.admin.* scopes. Defaults to "false" so the demo
--                    role cannot administer the server.
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
    ('gw.identity.user.read', 'Read and search community users', 'identity-discord'),
    ('gw.identity.roles.manage', 'Manage roles, permission grants and Discord role mappings', 'identity-discord'),
    ('gw.report.create', 'Submit a player report and read your own reports', 'reports'),
    ('gw.report.read', 'Read and close player reports', 'reports'),
    ('gw.audit.read', 'Read the audit log', 'gateway-admin'),
    ('gw.apikeys.manage', 'Create, rotate and revoke API keys', 'apikeys'),
    ('gw.notes.read', 'Read staff annotations', 'admin-notes'),
    ('gw.notes.write', 'Create staff annotations and edit your own', 'admin-notes'),
    ('gw.notes.manage', 'Edit and delete any staff annotation', 'admin-notes'),
    ('gw.store.catalog.read', 'Read the store product catalog', 'store'),
    ('gw.store.wallet.read', 'Read the authenticated user''s wallet balance', 'store'),
    ('gw.store.orders.read', 'Read the authenticated user''s orders', 'store'),
    ('gw.store.purchase', 'Buy store products', 'store'),
    ('gw.store.admin.wallets', 'Grant points to any wallet', 'store'),
    ('gw.store.admin.products', 'Manage the store product catalog', 'store'),
    ('gw.store.admin.orders.read', 'Read every store order', 'store'),
    ('gw.store.admin.orders.resolve', 'Refund or retry stuck store orders', 'store'),
    ('azeroth.info.public.read', 'Read public AzerothCore server information', 'azeroth-info'),
    ('azeroth.account.read', 'Read AzerothCore account information linked to a user', 'azeroth-account'),
    ('azeroth.account.manage', 'Create and manage AzerothCore accounts', 'azeroth-account'),
    ('azeroth.account.list', 'List AzerothCore login accounts', 'azeroth-account'),
    ('azeroth.account.link', 'Create and remove community user / AzerothCore account links', 'azeroth-account'),
    ('azeroth.account.self', 'Create and link your own AzerothCore account', 'azeroth-account'),
    ('azeroth.admin.claims.read', 'List pending account claims', 'azeroth-account'),
    ('azeroth.character.list', 'Read AzerothCore characters', 'azeroth-character'),
    ('azeroth.character.self', 'Read your own characters', 'azeroth-character'),
    ('azeroth.mail.send', 'Send in-game mail, items and money', 'azeroth-character'),
    ('azeroth.mail.self', 'Mail your own characters', 'azeroth-character'),
    ('azeroth.admin.mail.send', 'Send in-game mail to any character as staff', 'azeroth-character'),
    ('azeroth.leaderboard.read', 'Read character leaderboards', 'azeroth-character'),
    ('azeroth.item.list', 'Search the AzerothCore item catalog', 'azeroth-item'),
    ('azeroth.admin.accounts.ban', 'Ban and unban AzerothCore accounts', 'azeroth-admin'),
    ('azeroth.admin.accounts.gmlevel', 'Change AzerothCore account GM level', 'azeroth-admin'),
    ('azeroth.admin.players.read', 'List online players', 'azeroth-admin'),
    ('azeroth.admin.players.kick', 'Kick online players', 'azeroth-admin'),
    ('azeroth.admin.players.mute', 'Mute and unmute players', 'azeroth-admin'),
    ('azeroth.admin.characters.ban', 'Ban and unban characters', 'azeroth-admin'),
    ('azeroth.admin.announce', 'Broadcast announcements', 'azeroth-admin')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (name, description) VALUES
    (:'internal_role', 'Demo administrator')
ON CONFLICT (name) DO NOTHING;

-- Grant the role the registered permissions. Administrative scopes are only
-- granted when include_admin=true.
INSERT INTO role_permissions (role, permission)
SELECT :'internal_role', name FROM permissions
WHERE :'include_admin' = 'true'
   OR (name NOT LIKE 'azeroth.admin.%' AND name NOT LIKE 'gw.store.admin.%')
ON CONFLICT DO NOTHING;

-- Optional database mapping; equivalent static mappings can be provided with
-- ACGW_DISCORD_ROLE_MAPPINGS=<discord_role_id>=<internal_role>.
INSERT INTO discord_role_mappings (discord_role_id, role)
VALUES (:'discord_role_id', :'internal_role')
ON CONFLICT (discord_role_id) DO UPDATE
SET role = EXCLUDED.role,
    updated_at = now();

COMMIT;
