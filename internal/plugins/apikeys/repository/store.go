// Package repository implements the apikeys module's PostgreSQL persistence.
package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/apikeys/domain"
	apikeysrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/apikeys/repository/generated"
)

// ErrKeyNotFound is returned when an API key does not exist.
var ErrKeyNotFound = domain.ErrKeyNotFound

// Store persists and authenticates API keys.
type Store struct {
	q *apikeysrepo.Queries
}

// New creates a store over an open database pool.
func New(db *sql.DB) *Store {
	return &Store{q: apikeysrepo.New(db)}
}

// Create issues a new key and returns it with the one-time secret.
func (s *Store) Create(ctx context.Context, name string, permissions []string) (domain.APIKey, string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		raw, prefix, hash, err := generateKey()
		if err != nil {
			return domain.APIKey{}, "", err
		}
		row, err := s.q.InsertAPIKey(ctx, apikeysrepo.InsertAPIKeyParams{
			ID:          uuid.New(),
			Name:        name,
			KeyPrefix:   prefix,
			KeyHash:     hash,
			Permissions: strings.Join(permissions, " "),
		})
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return domain.APIKey{}, "", fmt.Errorf("repository: insert api key: %w", err)
		}
		return toDomain(row), raw, nil
	}
	return domain.APIKey{}, "", errors.New("repository: could not allocate a unique api key")
}

// List returns every key, newest first.
func (s *Store) List(ctx context.Context) ([]domain.APIKey, error) {
	rows, err := s.q.ListAPIKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: list api keys: %w", err)
	}
	keys := make([]domain.APIKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, toDomain(row))
	}
	return keys, nil
}

// Rotate issues a new secret for an existing key and returns it.
func (s *Store) Rotate(ctx context.Context, id uuid.UUID) (domain.APIKey, string, error) {
	existing, err := s.q.ListAPIKeys(ctx)
	if err != nil {
		return domain.APIKey{}, "", fmt.Errorf("repository: list api keys: %w", err)
	}
	var permissions []string
	found := false
	for _, row := range existing {
		if row.ID == id {
			permissions = splitPermissions(row.Permissions)
			found = true
			break
		}
	}
	if !found {
		return domain.APIKey{}, "", ErrKeyNotFound
	}
	raw, prefix, hash, err := generateKey()
	if err != nil {
		return domain.APIKey{}, "", err
	}
	row, err := s.q.RotateAPIKey(ctx, apikeysrepo.RotateAPIKeyParams{
		ID:          id,
		KeyPrefix:   prefix,
		KeyHash:     hash,
		Permissions: strings.Join(permissions, " "),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.APIKey{}, "", ErrKeyNotFound
	}
	if err != nil {
		return domain.APIKey{}, "", fmt.Errorf("repository: rotate api key: %w", err)
	}
	return toDomain(row), raw, nil
}

// Revoke deletes a key.
func (s *Store) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.q.DeleteAPIKey(ctx, id)
}

// Authenticate resolves a raw key to a principal carrying its explicit scopes.
func (s *Store) Authenticate(ctx context.Context, rawKey string) (auth.Principal, bool) {
	if !strings.HasPrefix(rawKey, "ak_") {
		return auth.Principal{}, false
	}
	row, err := s.q.GetAPIKeyByHash(ctx, hashKey(rawKey))
	if err != nil {
		return auth.Principal{}, false
	}
	_ = s.q.TouchAPIKey(ctx, row.ID)
	return auth.Principal{
		UserID:      row.ID,
		DiscordID:   "apikey:" + row.KeyPrefix,
		Permissions: splitPermissions(row.Permissions),
	}, true
}

func generateKey() (raw, prefix, hash string, err error) {
	prefixBytes := make([]byte, 4)
	if _, err := rand.Read(prefixBytes); err != nil {
		return "", "", "", fmt.Errorf("repository: generate key prefix: %w", err)
	}
	secretBytes := make([]byte, 24)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", "", fmt.Errorf("repository: generate key secret: %w", err)
	}
	prefix = hex.EncodeToString(prefixBytes)
	raw = "ak_" + prefix + "_" + hex.EncodeToString(secretBytes)
	return raw, prefix, hashKey(raw), nil
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func splitPermissions(value string) []string {
	fields := strings.Fields(value)
	if fields == nil {
		return []string{}
	}
	return fields
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toDomain(row apikeysrepo.ApiKey) domain.APIKey {
	key := domain.APIKey{
		ID:          row.ID,
		Name:        row.Name,
		KeyPrefix:   row.KeyPrefix,
		Permissions: splitPermissions(row.Permissions),
		CreatedAt:   row.CreatedAt,
	}
	if row.LastUsedAt.Valid {
		used := row.LastUsedAt.Time
		key.LastUsedAt = &used
	}
	return key
}
