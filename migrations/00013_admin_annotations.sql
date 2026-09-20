-- +goose Up

CREATE TABLE admin_annotations (
    id          uuid PRIMARY KEY,
    target_type text NOT NULL,
    target_id   text NOT NULL,
    author_id   uuid NOT NULL,
    body        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX admin_annotations_target_idx
    ON admin_annotations (target_type, target_id, created_at DESC);
CREATE INDEX admin_annotations_author_idx ON admin_annotations (author_id);

-- +goose Down

DROP TABLE admin_annotations;
