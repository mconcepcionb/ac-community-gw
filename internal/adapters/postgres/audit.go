package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
)

// AuditStore persists and reads audit entries in PostgreSQL.
type AuditStore struct {
	db *sql.DB
}

var (
	_ audit.Recorder = (*AuditStore)(nil)
	_ audit.Reader   = (*AuditStore)(nil)
)

// NewAuditStore creates an audit store over an open pool.
func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

// Record implements audit.Recorder.
func (s *AuditStore) Record(ctx context.Context, entry audit.Entry) error {
	metadata := []byte("{}")
	if len(entry.Metadata) > 0 {
		encoded, err := json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("postgres: marshal audit metadata: %w", err)
		}
		metadata = encoded
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO audit_log
    (occurred_at, actor_id, actor_discord_id, action, permission, target_type, target_id, result, request_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		entry.Timestamp, nullAuditActor(entry.ActorID), entry.ActorDiscordID, entry.Action,
		entry.Permission, entry.TargetType, entry.TargetID, string(entry.Result), entry.RequestID, metadata)
	if err != nil {
		return fmt.Errorf("postgres: insert audit: %w", err)
	}
	return nil
}

// List implements audit.Reader.
func (s *AuditStore) List(
	ctx context.Context,
	filter audit.ListFilter,
	limit, offset int,
) ([]audit.Entry, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	conditions := []string{"TRUE"}
	args := []any{}
	if filter.ActorID != "" {
		args = append(args, filter.ActorID)
		conditions = append(conditions, fmt.Sprintf("actor_id = $%d::uuid", len(args)))
	}
	if filter.Target != "" {
		args = append(args, filter.Target)
		conditions = append(conditions, fmt.Sprintf("target_id ILIKE '%%' || $%d || '%%'", len(args)))
	}
	if filter.TargetType != "" {
		args = append(args, filter.TargetType)
		conditions = append(conditions, fmt.Sprintf("target_type = $%d", len(args)))
	}
	if filter.TargetID != "" {
		args = append(args, filter.TargetID)
		conditions = append(conditions, fmt.Sprintf("target_id = $%d", len(args)))
	}
	if filter.Action != "" {
		args = append(args, filter.Action)
		conditions = append(conditions, fmt.Sprintf("action ILIKE '%%' || $%d || '%%'", len(args)))
	}
	if !filter.Since.IsZero() {
		args = append(args, filter.Since)
		conditions = append(conditions, fmt.Sprintf("occurred_at >= $%d", len(args)))
	}
	if !filter.Until.IsZero() {
		args = append(args, filter.Until)
		conditions = append(conditions, fmt.Sprintf("occurred_at <= $%d", len(args)))
	}
	args = append(args, limit)
	limitPlaceholder := len(args)
	args = append(args, offset)
	offsetPlaceholder := len(args)

	query := fmt.Sprintf(`SELECT occurred_at, actor_id, actor_discord_id, action, permission,
       target_type, target_id, result, request_id, metadata
FROM audit_log WHERE %s ORDER BY occurred_at DESC LIMIT $%d OFFSET $%d`,
		strings.Join(conditions, " AND "), limitPlaceholder, offsetPlaceholder)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: query audit: %w", err)
	}
	defer rows.Close()

	entries := make([]audit.Entry, 0)
	for rows.Next() {
		var (
			occurredAt     time.Time
			actorID        sql.NullString
			actorDiscordID string
			action         string
			permission     string
			targetType     string
			targetID       string
			result         string
			requestID      string
			metadata       []byte
		)
		if err := rows.Scan(&occurredAt, &actorID, &actorDiscordID, &action, &permission,
			&targetType, &targetID, &result, &requestID, &metadata); err != nil {
			return nil, fmt.Errorf("postgres: scan audit: %w", err)
		}
		entry := audit.Entry{
			Timestamp:      occurredAt,
			ActorDiscordID: actorDiscordID,
			Action:         action,
			Permission:     permission,
			TargetType:     targetType,
			TargetID:       targetID,
			Result:         audit.Result(result),
			RequestID:      requestID,
		}
		if actorID.Valid {
			if id, parseErr := uuid.Parse(actorID.String); parseErr == nil {
				entry.ActorID = id
			}
		}
		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &entry.Metadata)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate audit: %w", err)
	}
	return entries, nil
}

func nullAuditActor(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
