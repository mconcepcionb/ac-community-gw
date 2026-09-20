package azerothadmin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeExecutor struct {
	last   string
	result string
	err    error
}

func (f *fakeExecutor) Execute(_ context.Context, command string) (string, error) {
	f.last = command
	return f.result, f.err
}

func TestHandleBanAccount(t *testing.T) {
	executor := &fakeExecutor{result: "Bob is banned for 1 Day(s). Reason: cheating."}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts/Bob/ban",
		strings.NewReader(`{"duration":"1d","reason":"cheating"}`))
	req.SetPathValue("username", "Bob")
	rec := httptest.NewRecorder()
	plugin.handleBanAccount(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != ".ban account Bob 1d cheating" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleUnbanAccount(t *testing.T) {
	executor := &fakeExecutor{result: "Bob unbanned."}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts/Bob/unban", nil)
	req.SetPathValue("username", "Bob")
	rec := httptest.NewRecorder()
	plugin.handleUnbanAccount(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if executor.last != ".unban account Bob" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleSetGMLevel(t *testing.T) {
	executor := &fakeExecutor{result: "You change security level of account Bob to 3."}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/azeroth/accounts/Bob/gmlevel",
		strings.NewReader(`{"level":3,"realm":"-1"}`))
	req.SetPathValue("username", "Bob")
	rec := httptest.NewRecorder()
	plugin.handleSetGMLevel(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if executor.last != ".account set gmlevel Bob 3 -1" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleBanAccountValidation(t *testing.T) {
	plugin := New(&fakeExecutor{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts//ban",
		strings.NewReader(`{"duration":"1d"}`))
	rec := httptest.NewRecorder()
	plugin.handleBanAccount(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleBanAccountUpstreamFailure(t *testing.T) {
	plugin := New(&fakeExecutor{err: errors.New("boom")})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts/Bob/ban",
		strings.NewReader(`{"duration":"1d"}`))
	req.SetPathValue("username", "Bob")
	rec := httptest.NewRecorder()
	plugin.handleBanAccount(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", rec.Code)
	}
}
