-- +goose Up

CREATE TABLE azeroth_account_links (
    user_id          uuid PRIMARY KEY REFERENCES community_users (id) ON DELETE CASCADE,
    account_username text NOT NULL UNIQUE,
    account_id       integer,
    linked_at        timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

-- +goose Down

DROP TABLE azeroth_account_links;
