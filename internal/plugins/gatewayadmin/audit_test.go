package gatewayadmin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
)

type fakeAuditReader struct {
	filter  audit.ListFilter
	entries []audit.Entry
}

func (f *fakeAuditReader) List(_ context.Context, filter audit.ListFilter, _, _ int) ([]audit.Entry, error) {
	f.filter = filter
	return f.entries, nil
}

func TestAuditLogPassesTargetFilterAndMetadata(t *testing.T) {
	reader := &fakeAuditReader{entries: []audit.Entry{{
		Action:     "azeroth.accounts.ban",
		TargetType: "account",
		TargetID:   "Bob",
		Metadata:   map[string]any{"reason": "cheating"},
	}}}
	plugin := New(Config{AuditReader: reader})

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/admin/audit?target_type=account&target_id=Bob", nil)
	rec := httptest.NewRecorder()
	plugin.handleAuditLog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if reader.filter.TargetType != "account" || reader.filter.TargetID != "Bob" {
		t.Fatalf("filter = %+v", reader.filter)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"metadata"`) || !strings.Contains(body, `"reason"`) {
		t.Fatalf("metadata missing from response: %s", body)
	}
}
