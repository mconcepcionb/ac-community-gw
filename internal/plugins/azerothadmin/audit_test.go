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
