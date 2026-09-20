-- +goose Up

CREATE TABLE community_users (
    id           uuid PRIMARY KEY,
    discord_id   text NOT NULL UNIQUE,
    display_name text NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE discord_identities (
    user_id    uuid PRIMARY KEY REFERENCES community_users (id) ON DELETE CASCADE,
    discord_id text NOT NULL UNIQUE,
    username   text NOT NULL DEFAULT '',
    global_name text NOT NULL DEFAULT '',
    avatar     text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id         text PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES community_users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

-- +goose Down

DROP TABLE sessions;
DROP TABLE discord_identities;
DROP TABLE community_users;
