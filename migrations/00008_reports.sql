-- +goose Up

CREATE TABLE community_reports (
    id          uuid PRIMARY KEY,
    reporter_id uuid NOT NULL,
    target      text NOT NULL,
    category    text NOT NULL,
    message     text NOT NULL,
    status      text NOT NULL DEFAULT 'open',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX community_reports_reporter_idx ON community_reports (reporter_id);
CREATE INDEX community_reports_status_idx ON community_reports (status);

-- +goose Down

DROP TABLE community_reports;
