-- +goose Up

CREATE TABLE api_keys (
    id           uuid PRIMARY KEY,
    name         text NOT NULL,
    key_prefix   text NOT NULL,
    key_hash     text NOT NULL UNIQUE,
    permissions  text NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz
);

CREATE INDEX api_keys_key_prefix_idx ON api_keys (key_prefix);

-- +goose Down

DROP TABLE api_keys;
