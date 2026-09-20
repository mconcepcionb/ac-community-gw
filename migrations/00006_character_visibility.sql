-- +goose Up

CREATE TABLE character_visibility (
    character_name text PRIMARY KEY,
    user_id        uuid NOT NULL,
    public         boolean NOT NULL DEFAULT false,
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX character_visibility_user_idx ON character_visibility (user_id);
CREATE INDEX character_visibility_public_idx ON character_visibility (public) WHERE public;

-- +goose Down

DROP TABLE character_visibility;
