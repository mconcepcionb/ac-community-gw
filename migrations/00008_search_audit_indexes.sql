-- +goose Up

-- Trigram search for community users. Leading-wildcard ILIKE cannot use a
-- btree index; a GIN trigram index supports the substring search.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX community_users_display_name_trgm_idx
    ON community_users USING gin (display_name gin_trgm_ops);
CREATE INDEX discord_identities_username_trgm_idx
    ON discord_identities USING gin (username gin_trgm_ops);
CREATE INDEX discord_identities_global_name_trgm_idx
    ON discord_identities USING gin (global_name gin_trgm_ops);

-- Audit filters.
CREATE INDEX audit_log_actor_id_idx ON audit_log (actor_id);
CREATE INDEX audit_log_action_idx ON audit_log (action);
CREATE INDEX audit_log_target_type_idx ON audit_log (target_type);

-- Retention: audit_log grows unbounded. The table is indexed for time-based
-- range deletes so an operator (or a future janitor) can prune it safely; this
-- migration does not delete any history.
CREATE INDEX audit_log_occurred_at_brin_idx
    ON audit_log USING brin (occurred_at);

-- +goose Down

DROP INDEX IF EXISTS audit_log_occurred_at_brin_idx;
DROP INDEX IF EXISTS audit_log_target_type_idx;
DROP INDEX IF EXISTS audit_log_action_idx;
DROP INDEX IF EXISTS audit_log_actor_id_idx;
DROP INDEX IF EXISTS discord_identities_global_name_trgm_idx;
DROP INDEX IF EXISTS discord_identities_username_trgm_idx;
DROP INDEX IF EXISTS community_users_display_name_trgm_idx;
-- pg_trgm is shared; leave the extension installed.
