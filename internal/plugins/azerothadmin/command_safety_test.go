package azerothadmin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminCommandInjectionRejected(t *testing.T) {
	tests := []struct {
		name     string
		handler  func(*Plugin, http.ResponseWriter, *http.Request)
		method   string
		path     string
		pathUser string
		body     string
	}{
		{
			name:     "ban injection in username",
			handler:  (*Plugin).handleBanAccount,
			method:   http.MethodPost,
			path:     "/api/v1/azeroth/accounts/Bob/ban",
			pathUser: "Bob\n.server shutdown",
			body:     `{"duration":"1d","reason":"x"}`,
		},
		{
			name:     "ban injection in duration",
			handler:  (*Plugin).handleBanAccount,
			method:   http.MethodPost,
			path:     "/api/v1/azeroth/accounts/Bob/ban",
			pathUser: "Bob",
			body:     `{"duration":"1d; .server shutdown","reason":"x"}`,
		},
		{
			name:     "unban injection in username",
			handler:  (*Plugin).handleUnbanAccount,
			method:   http.MethodPost,
			path:     "/api/v1/azeroth/accounts/Bob/unban",
			pathUser: "Bob;x",
		},
		{
			name:     "gmlevel injection in realm",
			handler:  (*Plugin).handleSetGMLevel,
			method:   http.MethodPut,
			path:     "/api/v1/azeroth/accounts/Bob/gmlevel",
			pathUser: "Bob",
			body:     `{"level":3,"realm":"-1; .server shutdown"}`,
		},
		{
			name:     "gmlevel out of range",
			handler:  (*Plugin).handleSetGMLevel,
			method:   http.MethodPut,
			path:     "/api/v1/azeroth/accounts/Bob/gmlevel",
			pathUser: "Bob",
			body:     `{"level":99,"realm":"-1"}`,
		},
		{
			name:     "mute invalid duration",
			handler:  (*Plugin).handleMute,
			method:   http.MethodPost,
			path:     "/api/v1/azeroth/players/Thrall/mute",
			pathUser: "",
			body:     `{"duration":"tomorrow","reason":"x"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &fakeExecutor{result: "should not run"}
			plugin := New(executor)
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.pathUser != "" {
				req.SetPathValue("username", tt.pathUser)
				req.SetPathValue("name", tt.pathUser)
			}
			if strings.Contains(tt.path, "/players/") {
				req.SetPathValue("name", "Thrall")
			}
			rec := httptest.NewRecorder()
			tt.handler(plugin, rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if executor.last != "" {
				t.Fatalf("injection reached executor: %q", executor.last)
			}
		})
	}
}

// TestBanReasonIsSanitized asserts that free-text reasons are neutralized rather
// than rejected, and never introduce a line break.
func TestBanReasonIsSanitized(t *testing.T) {
	executor := &fakeExecutor{result: "banned"}
	plugin := New(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts/Bob/ban",
		strings.NewReader("{\"duration\":\"1d\",\"reason\":\"bad\\n.server shutdown\"}"))
	req.SetPathValue("username", "Bob")
	rec := httptest.NewRecorder()
	plugin.handleBanAccount(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.ContainsAny(executor.last, "\r\n") {
		t.Fatalf("command contains a line break: %q", executor.last)
	}
}
