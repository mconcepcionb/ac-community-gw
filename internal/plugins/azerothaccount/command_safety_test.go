package azerothaccount

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCommandInjectionRejected asserts that no injection payload reaches the
// AzerothCore executor through the account endpoints.
func TestCommandInjectionRejected(t *testing.T) {
	tests := []struct {
		name     string
		handler  func(*Plugin, http.ResponseWriter, *http.Request)
		method   string
		path     string
		pathUser string
		body     string
	}{
		{
			name:    "create injection in username",
			handler: (*Plugin).handleCreateAccount,
			method:  http.MethodPost,
			path:    "/api/v1/azeroth/accounts",
			body:    `{"username":"Alice\n.server shutdown","password":"secret"}`,
		},
		{
			name:    "create injection in password",
			handler: (*Plugin).handleCreateAccount,
			method:  http.MethodPost,
			path:    "/api/v1/azeroth/accounts",
			body:    `{"username":"Alice","password":"secret extra"}`,
		},
		{
			name:    "create injection in email",
			handler: (*Plugin).handleCreateAccount,
			method:  http.MethodPost,
			path:    "/api/v1/azeroth/accounts",
			body:    `{"username":"Alice","password":"secret","email":"a@b.c\n.server shutdown"}`,
		},
		{
			name:     "change password injection in path username",
			handler:  (*Plugin).handleChangePassword,
			method:   http.MethodPost,
			path:     "/api/v1/azeroth/accounts/Alice%20bar/password",
			pathUser: "Alice bar",
			body:     `{"password":"secret"}`,
		},
		{
			name:     "set email injection in path username",
			handler:  (*Plugin).handleSetEmail,
			method:   http.MethodPut,
			path:     "/api/v1/azeroth/accounts/Alice%0A.server/password",
			pathUser: "Alice\n.server",
			body:     `{"email":"a@b.co"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &fakeExecutor{result: "should not run"}
			plugin := New(Config{Executor: executor})
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.pathUser != "" {
				req.SetPathValue("username", tt.pathUser)
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
