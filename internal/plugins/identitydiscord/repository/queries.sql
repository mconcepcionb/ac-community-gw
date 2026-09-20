-- name: CreateCommunityUser :one
INSERT INTO community_users (id, discord_id, display_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpsertCommunityUser :one
INSERT INTO community_users (id, discord_id, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (discord_id) DO UPDATE
SET display_name = EXCLUDED.display_name,
    updated_at = now()
RETURNING *, (xmax = 0) AS inserted;

-- name: GetCommunityUserByID :one
SELECT * FROM community_users WHERE id = $1;

-- name: GetCommunityUserByDiscordID :one
SELECT * FROM community_users WHERE discord_id = $1;

-- name: GetUserProfile :one
SELECT cu.id,
       cu.discord_id,
       cu.display_name,
       cu.created_at,
       COALESCE(di.username, '') AS username,
       COALESCE(di.global_name, '') AS global_name,
       COALESCE(di.avatar, '') AS avatar
FROM community_users cu
LEFT JOIN discord_identities di ON di.user_id = cu.id
WHERE cu.id = $1;

-- name: UpdateCommunityUserRoles :exec
UPDATE community_users
SET roles = $2::text[],
    roles_synced_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: UpsertDiscordIdentity :one
INSERT INTO discord_identities (user_id, discord_id, username, global_name, avatar)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id) DO UPDATE
SET username = EXCLUDED.username,
    global_name = EXCLUDED.global_name,
    avatar = EXCLUDED.avatar,
    updated_at = now()
RETURNING *;

-- name: GetDiscordIdentityByUserID :one
SELECT * FROM discord_identities WHERE user_id = $1;

-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at, roles)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSession :one
SELECT s.id, s.user_id, s.created_at, s.expires_at, s.roles, cu.discord_id
FROM sessions s
JOIN community_users cu ON cu.id = s.user_id
WHERE s.id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < now();

-- name: ListInternalRolesForDiscordRoles :many
SELECT DISTINCT role
FROM discord_role_mappings
WHERE discord_role_id = ANY($1::text[]);

-- name: UpsertDiscordRoleMapping :exec
INSERT INTO discord_role_mappings (discord_role_id, role)
VALUES ($1, $2)
ON CONFLICT (discord_role_id) DO UPDATE
SET role = EXCLUDED.role,
    updated_at = now();

-- name: ListRolePermissions :many
SELECT role, permission FROM role_permissions;

-- name: UpsertPermission :exec
INSERT INTO permissions (name, description, owner)
VALUES ($1, $2, $3)
ON CONFLICT (name) DO UPDATE
SET description = EXCLUDED.description,
    owner = EXCLUDED.owner;

-- name: UpsertRole :exec
INSERT INTO roles (name) VALUES ($1)
ON CONFLICT (name) DO NOTHING;

-- name: GrantRolePermission :exec
INSERT INTO role_permissions (role, permission)
VALUES ($1, $2)
ON CONFLICT (role, permission) DO NOTHING;

-- name: RevokeRolePermission :exec
DELETE FROM role_permissions WHERE role = $1 AND permission = $2;

-- name: ListDiscordRoleMappings :many
SELECT * FROM discord_role_mappings ORDER BY discord_role_id;

-- name: DeleteDiscordRoleMapping :exec
DELETE FROM discord_role_mappings WHERE discord_role_id = $1;

-- name: CreateOAuthState :exec
INSERT INTO oauth_states (state, code_verifier, return_to, expires_at)
VALUES ($1, $2, $3, $4);

-- name: ConsumeOAuthState :one
DELETE FROM oauth_states
WHERE state = $1 AND expires_at > now()
RETURNING state, code_verifier, return_to, created_at, expires_at;

-- name: DeleteExpiredOAuthStates :exec
DELETE FROM oauth_states WHERE expires_at < now();

-- name: ListCommunityUsers :many
SELECT cu.id,
       cu.discord_id,
       cu.display_name,
       cu.created_at,
       COALESCE(di.username, '') AS username,
       COALESCE(di.global_name, '') AS global_name
FROM community_users cu
LEFT JOIN discord_identities di ON di.user_id = cu.id
WHERE ($1 = '' OR cu.discord_id ILIKE $2 OR cu.display_name ILIKE $2
       OR di.username ILIKE $2 OR di.global_name ILIKE $2)
ORDER BY cu.created_at DESC
LIMIT $3 OFFSET $4;

-- name: ResolveUserByName :many
SELECT cu.id,
       cu.discord_id,
       cu.display_name,
       cu.created_at,
       COALESCE(di.username, '') AS username,
       COALESCE(di.global_name, '') AS global_name
FROM community_users cu
LEFT JOIN discord_identities di ON di.user_id = cu.id
WHERE lower(cu.display_name) = lower(sqlc.arg(name))
   OR lower(di.username) = lower(sqlc.arg(name))
   OR lower(di.global_name) = lower(sqlc.arg(name))
   OR cu.discord_id = sqlc.arg(name)
ORDER BY cu.created_at DESC
LIMIT 50;
