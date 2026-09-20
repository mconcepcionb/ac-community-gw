// Package repository implements the reports module's PostgreSQL persistence.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/reports/domain"
	reportsrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/reports/repository/generated"
)

// Store persists community reports.
type Store struct {
	q *reportsrepo.Queries
}

// New creates a store over an open database pool.
func New(db *sql.DB) *Store {
	return &Store{q: reportsrepo.New(db)}
}

// Create inserts a new open report.
func (s *Store) Create(
	ctx context.Context,
	reporterID uuid.UUID,
	target, category, message string,
) (domain.Report, error) {
	row, err := s.q.InsertReport(ctx, reportsrepo.InsertReportParams{
		ID:         uuid.New(),
		ReporterID: reporterID,
		Target:     target,
		Category:   category,
		Message:    message,
	})
	if err != nil {
		return domain.Report{}, fmt.Errorf("repository: insert report: %w", err)
	}
	return toDomain(row), nil
}

// ByReporter lists a user's own reports, newest first.
func (s *Store) ByReporter(
	ctx context.Context,
	reporterID uuid.UUID,
	limit, offset int,
) ([]domain.Report, error) {
	rows, err := s.q.ListReportsByReporter(ctx, reportsrepo.ListReportsByReporterParams{
		ReporterID: reporterID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("repository: list reports by reporter: %w", err)
	}
	return toDomainSlice(rows), nil
}

// List lists all reports, newest first, optionally filtered by status.
func (s *Store) List(
	ctx context.Context,
	status string,
	limit, offset int,
) ([]domain.Report, error) {
	rows, err := s.q.ListReports(ctx, reportsrepo.ListReportsParams{
		Limit:   int32(limit),
		Offset:  int32(offset),
		Column3: status,
	})
	if err != nil {
		return nil, fmt.Errorf("repository: list reports: %w", err)
	}
	return toDomainSlice(rows), nil
}

// Close marks a report closed, or returns domain.ErrReportNotFound.
func (s *Store) Close(ctx context.Context, id uuid.UUID) (domain.Report, error) {
	row, err := s.q.CloseReport(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Report{}, domain.ErrReportNotFound
	}
	if err != nil {
		return domain.Report{}, fmt.Errorf("repository: close report: %w", err)
	}
	return toDomain(row), nil
}

func toDomainSlice(rows []reportsrepo.CommunityReport) []domain.Report {
	reports := make([]domain.Report, 0, len(rows))
	for _, row := range rows {
		reports = append(reports, toDomain(row))
	}
	return reports
}

func toDomain(row reportsrepo.CommunityReport) domain.Report {
	return domain.Report{
		ID:         row.ID,
		ReporterID: row.ReporterID,
		Target:     row.Target,
		Category:   row.Category,
		Message:    row.Message,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt,
	}
}
