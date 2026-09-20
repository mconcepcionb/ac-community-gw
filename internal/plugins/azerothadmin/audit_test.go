package azerothadmin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
)

type recordingAudit struct {
	entries []audit.Entry
}

func (r *recordingAudit) Record(_ context.Context, entry audit.Entry) error {
	r.entries = append(r.entries, entry)
	return nil
}

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
	plugin := New(&fakeExecutor{}, WithAuditReader(reader))

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

func TestBanAccountRecordsAudit(t *testing.T) {
	executor := &fakeExecutor{result: "banned"}
	recorder := &recordingAudit{}
	plugin := New(executor, WithAudit(recorder))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts/Bob/ban",
		strings.NewReader(`{"duration":"1d","reason":"cheating"}`))
	req.SetPathValue("username", "Bob")
	rec := httptest.NewRecorder()
	plugin.handleBanAccount(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(recorder.entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(recorder.entries))
	}
	entry := recorder.entries[0]
	if entry.Action != "azeroth.accounts.ban" || entry.Result != audit.ResultSuccess || entry.TargetID != "Bob" {
		t.Fatalf("entry = %+v", entry)
	}
}
