-- +goose Up

CREATE TABLE account_claims (
    user_id          uuid PRIMARY KEY,
    account_username text NOT NULL,
    code_hash        text NOT NULL,
    expires_at       timestamptz NOT NULL,
    attempts         int NOT NULL DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

-- +goose Down

DROP TABLE account_claims;
