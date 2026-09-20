// Package repository implements the admin notes module's PostgreSQL persistence.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/adminnotes/domain"
	adminnotesrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/adminnotes/repository/generated"
)

// Store persists admin annotations.
type Store struct {
	q *adminnotesrepo.Queries
}

// New creates a store over an open database pool.
func New(db *sql.DB) *Store {
	return &Store{q: adminnotesrepo.New(db)}
}

// Create inserts a new annotation.
func (s *Store) Create(
	ctx context.Context,
	targetType, targetID string,
	authorID uuid.UUID,
	body string,
) (domain.Annotation, error) {
	row, err := s.q.InsertAnnotation(ctx, adminnotesrepo.InsertAnnotationParams{
		ID:         uuid.New(),
		TargetType: targetType,
		TargetID:   targetID,
		AuthorID:   authorID,
		Body:       body,
	})
	if err != nil {
		return domain.Annotation{}, fmt.Errorf("repository: insert annotation: %w", err)
	}
	return toDomain(row), nil
}

// ByTarget lists the annotations of one target, newest first.
func (s *Store) ByTarget(
	ctx context.Context,
	targetType, targetID string,
	limit, offset int,
) ([]domain.Annotation, error) {
	rows, err := s.q.ListAnnotations(ctx, adminnotesrepo.ListAnnotationsParams{
		TargetType: targetType,
		TargetID:   targetID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("repository: list annotations: %w", err)
	}
	annotations := make([]domain.Annotation, 0, len(rows))
	for _, row := range rows {
		annotations = append(annotations, toDomain(row))
	}
	return annotations, nil
}

// Get returns an annotation, or domain.ErrAnnotationNotFound.
func (s *Store) Get(ctx context.Context, id uuid.UUID) (domain.Annotation, error) {
	row, err := s.q.GetAnnotation(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Annotation{}, domain.ErrAnnotationNotFound
	}
	if err != nil {
		return domain.Annotation{}, fmt.Errorf("repository: get annotation: %w", err)
	}
	return toDomain(row), nil
}

// Update replaces an annotation's body, or returns domain.ErrAnnotationNotFound.
func (s *Store) Update(ctx context.Context, id uuid.UUID, body string) (domain.Annotation, error) {
	row, err := s.q.UpdateAnnotation(ctx, adminnotesrepo.UpdateAnnotationParams{ID: id, Body: body})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Annotation{}, domain.ErrAnnotationNotFound
	}
	if err != nil {
		return domain.Annotation{}, fmt.Errorf("repository: update annotation: %w", err)
	}
	return toDomain(row), nil
}

// Delete removes an annotation, or returns domain.ErrAnnotationNotFound.
func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := s.q.DeleteAnnotation(ctx, id)
	if err != nil {
		return fmt.Errorf("repository: delete annotation: %w", err)
	}
	if affected == 0 {
		return domain.ErrAnnotationNotFound
	}
	return nil
}

func toDomain(row adminnotesrepo.AdminAnnotation) domain.Annotation {
	return domain.Annotation{
		ID:         row.ID,
		TargetType: row.TargetType,
		TargetID:   row.TargetID,
		AuthorID:   row.AuthorID,
		Body:       row.Body,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
