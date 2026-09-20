-- +goose Up

CREATE TABLE roles (
    name        text PRIMARY KEY,
    description text NOT NULL DEFAULT ''
);

CREATE TABLE permissions (
    name        text PRIMARY KEY,
    description text NOT NULL DEFAULT '',
    owner       text NOT NULL DEFAULT ''
);

CREATE TABLE role_permissions (
    role       text NOT NULL REFERENCES roles (name) ON DELETE CASCADE,
    permission text NOT NULL REFERENCES permissions (name) ON DELETE CASCADE,
    PRIMARY KEY (role, permission)
);

CREATE TABLE audit_log (
    id               bigserial PRIMARY KEY,
    occurred_at      timestamptz NOT NULL DEFAULT now(),
    actor_id         uuid,
    actor_discord_id text NOT NULL DEFAULT '',
    action           text NOT NULL,
    permission       text NOT NULL DEFAULT '',
    target_type      text NOT NULL DEFAULT '',
    target_id        text NOT NULL DEFAULT '',
    result           text NOT NULL,
    request_id       text NOT NULL DEFAULT '',
    metadata         jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX audit_log_occurred_at_idx ON audit_log (occurred_at DESC);

-- +goose Down

DROP TABLE audit_log;
DROP TABLE role_permissions;
DROP TABLE permissions;
DROP TABLE roles;
