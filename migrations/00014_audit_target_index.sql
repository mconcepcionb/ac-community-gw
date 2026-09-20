-- +goose Up

-- The per-entity history panel queries (target_type, target_id) ordered by
-- time. A composite index replaces the single-column target_type index.
DROP INDEX IF EXISTS audit_log_target_type_idx;
CREATE INDEX audit_log_target_idx
    ON audit_log (target_type, target_id, occurred_at DESC);

-- +goose Down

DROP INDEX IF EXISTS audit_log_target_idx;
CREATE INDEX audit_log_target_type_idx ON audit_log (target_type);
