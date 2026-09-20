// Package repository implements the PostgreSQL persistence owned by the
// azeroth-character plugin: the per-character public visibility flag.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	azerothcharacterrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothcharacter/repository/generated"
)

// ErrVisibilityNotFound is returned when a character has no visibility row.
var ErrVisibilityNotFound = errors.New("repository: character visibility not found")

// Store persists character visibility flags.
type Store struct {
	q *azerothcharacterrepo.Queries
}

// New creates a store over an open database pool.
func New(db *sql.DB) *Store {
	return &Store{q: azerothcharacterrepo.New(db)}
}

// SetVisibility upserts the public flag for a character owned by userID.
func (s *Store) SetVisibility(
	ctx context.Context,
	characterName string,
	userID uuid.UUID,
	public bool,
) (bool, error) {
	row, err := s.q.UpsertVisibility(ctx, azerothcharacterrepo.UpsertVisibilityParams{
		CharacterName: characterName,
		UserID:        userID,
		Public:        public,
	})
	if err != nil {
		return false, fmt.Errorf("repository: upsert visibility: %w", err)
	}
	return row.Public, nil
}

// GetVisibility returns the public flag of a character.
func (s *Store) GetVisibility(ctx context.Context, characterName string) (bool, error) {
	row, err := s.q.GetVisibility(ctx, characterName)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrVisibilityNotFound
	}
	if err != nil {
		return false, fmt.Errorf("repository: get visibility: %w", err)
	}
	return row.Public, nil
}

// ListByUser returns the public flag per character for a community user.
func (s *Store) ListByUser(ctx context.Context, userID uuid.UUID) (map[string]bool, error) {
	rows, err := s.q.ListVisibilityByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list visibility: %w", err)
	}
	out := make(map[string]bool, len(rows))
	for _, row := range rows {
		out[row.CharacterName] = row.Public
	}
	return out, nil
}

// PublicNames returns the names of characters opted into public boards.
func (s *Store) PublicNames(ctx context.Context) ([]string, error) {
	names, err := s.q.ListPublicCharacterNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: list public names: %w", err)
	}
	return names, nil
}
