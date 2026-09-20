// Package repository implements the azeroth-account module's persistence.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
	azerothaccountrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/repository/generated"
)

// Store implements the azeroth-account link persistence.
type Store struct {
	q *azerothaccountrepo.Queries
}

// New creates a store over an open database pool.
func New(db *sql.DB) *Store {
	return &Store{q: azerothaccountrepo.New(db)}
}

// Upsert creates or replaces a user's account link.
func (s *Store) Upsert(ctx context.Context, userID uuid.UUID, accountUsername string, accountID *int64) (domain.Link, error) {
	row, err := s.q.UpsertAccountLink(ctx, azerothaccountrepo.UpsertAccountLinkParams{
		UserID:          userID,
		AccountUsername: accountUsername,
		AccountID:       nullInt64(accountID),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Link{}, domain.ErrAccountAlreadyLinked
		}
		return domain.Link{}, fmt.Errorf("repository: upsert account link: %w", err)
	}
	return toDomain(row), nil
}

// Get returns a user's link or domain.ErrLinkNotFound.
func (s *Store) Get(ctx context.Context, userID uuid.UUID) (domain.Link, error) {
	row, err := s.q.GetAccountLinkByUserID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Link{}, domain.ErrLinkNotFound
	}
	if err != nil {
		return domain.Link{}, fmt.Errorf("repository: get account link: %w", err)
	}
	return toDomain(row), nil
}

// Delete removes a user's link. Unknown users are not an error.
func (s *Store) Delete(ctx context.Context, userID uuid.UUID) error {
	return s.q.DeleteAccountLink(ctx, userID)
}

// List returns every account link, newest first.
func (s *Store) List(ctx context.Context) ([]domain.Link, error) {
	rows, err := s.q.ListAccountLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: list account links: %w", err)
	}
	links := make([]domain.Link, 0, len(rows))
	for _, row := range rows {
		links = append(links, toDomain(row))
	}
	return links, nil
}

func toDomain(row azerothaccountrepo.AzerothAccountLink) domain.Link {
	var accountID *int64
	if row.AccountID.Valid {
		value := row.AccountID.Int64
		accountID = &value
	}
	return domain.Link{
		UserID:          row.UserID,
		AccountUsername: row.AccountUsername,
		AccountID:       accountID,
		LinkedAt:        row.LinkedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func nullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}
