package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/commands"
)

func TestRequestIDValidation(t *testing.T) {
	server, _, _ := testServer(t)

	valid := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	valid.Header.Set("X-Request-Id", "abc-123_ok")
	validRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(validRec, valid)
	if got := validRec.Header().Get("X-Request-Id"); got != "abc-123_ok" {
		t.Fatalf("valid id = %q", got)
	}

	invalid := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	invalid.Header.Set("X-Request-Id", "bad id with spaces")
	invalidRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(invalidRec, invalid)
	got := invalidRec.Header().Get("X-Request-Id")
	if got == "bad id with spaces" || got == "" {
		t.Fatalf("invalid id not replaced: %q", got)
	}
}

func TestStatusRecorderUnwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	recorder := &statusRecorder{ResponseWriter: rec}
	if recorder.Unwrap() != rec {
		t.Fatal("Unwrap did not return the underlying writer")
	}
}

func TestStatusForErrorMapping(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{azerothdb.ErrUnavailable, http.StatusServiceUnavailable},
		{azerothcore.ErrNotConfigured, http.StatusServiceUnavailable},
		{commands.ErrNotFound, http.StatusNotFound},
	}
	for _, tc := range cases {
		if got := StatusForError(tc.err).Status; got != tc.want {
			t.Fatalf("StatusForError(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}
