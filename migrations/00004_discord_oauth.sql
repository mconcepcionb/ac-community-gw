-- +goose Up

ALTER TABLE sessions
    ADD COLUMN roles text[] NOT NULL DEFAULT '{}';

ALTER TABLE community_users
    ADD COLUMN roles text[] NOT NULL DEFAULT '{}',
    ADD COLUMN roles_synced_at timestamptz;

CREATE TABLE discord_role_mappings (
    discord_role_id text PRIMARY KEY,
    role            text NOT NULL REFERENCES roles (name) ON DELETE CASCADE,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE oauth_states (
    state         text PRIMARY KEY,
    code_verifier text NOT NULL DEFAULT '',
    return_to     text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    expires_at    timestamptz NOT NULL
);

CREATE INDEX oauth_states_expires_at_idx ON oauth_states (expires_at);

-- +goose Down

DROP TABLE oauth_states;
DROP TABLE discord_role_mappings;
ALTER TABLE community_users DROP COLUMN roles_synced_at, DROP COLUMN roles;
ALTER TABLE sessions DROP COLUMN roles;
