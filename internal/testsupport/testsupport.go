// Package testsupport provides small helpers shared by package tests.
//
// It must not import any plugin: it only knows the standard library and core
// packages, so it cannot bridge the architecture boundaries. The architecture
// test enforces this.
package testsupport

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
)

// NewRecorder returns a response recorder.
func NewRecorder() *httptest.ResponseRecorder { return httptest.NewRecorder() }

// NewRequest builds a request with an optional JSON body.
func NewRequest(method, target, body string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// WithPrincipal authenticates the request as userID.
func WithPrincipal(req *http.Request, userID uuid.UUID) *http.Request {
	return req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{
		UserID:    userID,
		DiscordID: "42",
	}))
}

// SetPathValue sets a routing path parameter and returns the request.
func SetPathValue(req *http.Request, key, value string) *http.Request {
	req.SetPathValue(key, value)
	return req
}

// DecodeJSON decodes the recorder body into v, failing the test on error.
func DecodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode JSON body %q: %v", rec.Body.String(), err)
	}
}

// AssertStatus fails the test unless the recorder status matches.
func AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, want, rec.Body.String())
	}
}
