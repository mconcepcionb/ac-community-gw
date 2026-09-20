package azerothadmin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleKick(t *testing.T) {
	executor := &fakeExecutor{result: "Player kicked."}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/players/Thrall/kick",
		strings.NewReader(`{"reason":"afk"}`))
	req.SetPathValue("name", "Thrall")
	rec := httptest.NewRecorder()
	plugin.handleKick(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != ".kick Thrall afk" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleMute(t *testing.T) {
	executor := &fakeExecutor{result: "Player muted."}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/players/Thrall/mute",
		strings.NewReader(`{"duration":"2h","reason":"spam"}`))
	req.SetPathValue("name", "Thrall")
	rec := httptest.NewRecorder()
	plugin.handleMute(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if executor.last != ".mute Thrall 2h spam" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleMuteDefaultsDurationAndSanitizesReason(t *testing.T) {
	executor := &fakeExecutor{result: "muted"}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/players/Thrall/mute",
		strings.NewReader(`{"reason":"bad \"boy\"\nline"}`))
	req.SetPathValue("name", "Thrall")
	rec := httptest.NewRecorder()
	plugin.handleMute(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if executor.last != ".mute Thrall 1d bad boy line" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleBanCharacter(t *testing.T) {
	executor := &fakeExecutor{result: "banned"}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/characters/Thrall/ban",
		strings.NewReader(`{"duration":"1d","reason":"cheating"}`))
	req.SetPathValue("name", "Thrall")
	rec := httptest.NewRecorder()
	plugin.handleBanCharacter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if executor.last != ".ban character Thrall 1d cheating" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleUnbanCharacter(t *testing.T) {
	executor := &fakeExecutor{result: "unbanned"}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/characters/Thrall/unban", nil)
	req.SetPathValue("name", "Thrall")
	rec := httptest.NewRecorder()
	plugin.handleUnbanCharacter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if executor.last != ".unban character Thrall" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleKickInvalidName(t *testing.T) {
	plugin := New(&fakeExecutor{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/players/x/kick", nil)
	req.SetPathValue("name", "x")
	rec := httptest.NewRecorder()
	plugin.handleKick(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleAnnounce(t *testing.T) {
	executor := &fakeExecutor{result: "announced"}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/announce",
		strings.NewReader(`{"message":"Server restart in 5 minutes"}`))
	rec := httptest.NewRecorder()
	plugin.handleAnnounce(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != ".announce Server restart in 5 minutes" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleAnnounceEmpty(t *testing.T) {
	plugin := New(&fakeExecutor{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/announce",
		strings.NewReader(`{"message":"  "}`))
	rec := httptest.NewRecorder()
	plugin.handleAnnounce(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
}
