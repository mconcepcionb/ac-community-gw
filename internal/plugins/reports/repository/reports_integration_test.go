//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/reports/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestReportLifecycle(t *testing.T) {
	db := integration.Postgres(t)
	ctx := context.Background()
	store := New(db)

	reporterID := uuid.New()
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM community_reports WHERE reporter_id = $1`, reporterID) })

	report, err := store.Create(ctx, reporterID, "Thrall", "abuse", "spam")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if report.Status != domain.StatusOpen || report.ID == uuid.Nil {
		t.Fatalf("created = %+v", report)
	}

	mine, err := store.ByReporter(ctx, reporterID, 10, 0)
	if err != nil || len(mine) != 1 || mine[0].ID != report.ID {
		t.Fatalf("ByReporter = %+v, %v", mine, err)
	}

	open, err := store.List(ctx, domain.StatusOpen, 50, 0)
	if err != nil || !containsReport(open, report.ID) {
		t.Fatalf("List open = %+v, %v", open, err)
	}

	closed, err := store.Close(ctx, report.ID)
	if err != nil || closed.Status != domain.StatusClosed {
		t.Fatalf("Close = %+v, %v", closed, err)
	}
	if _, err := store.Close(ctx, uuid.New()); !errors.Is(err, domain.ErrReportNotFound) {
		t.Fatalf("Close missing = %v, want ErrReportNotFound", err)
	}
}

func containsReport(reports []domain.Report, id uuid.UUID) bool {
	for _, report := range reports {
		if report.ID == id {
			return true
		}
	}
	return false
}
