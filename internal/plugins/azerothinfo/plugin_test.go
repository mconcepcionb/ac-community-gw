package azerothinfo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeExecutor struct {
	output string
	err    error
}

func (f fakeExecutor) Execute(context.Context, string) (string, error) {
	return f.output, f.err
}

func TestHandleStatusReturnsTypedResponse(t *testing.T) {
	output := "AzerothCore rev. fake\n" +
		"Connected players: 3. Characters in world: 5.\n" +
		"Connection peak: 10.\n" +
		"Server uptime: 1 Hour(s) 2 Minute(s)"
	plugin := New(fakeExecutor{output: output})

	rec := httptest.NewRecorder()
	plugin.handleStatus(rec, httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/info/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	want := `{"output":"AzerothCore rev. fake\nConnected players: 3. Characters in world: 5.\nConnection peak: 10.\nServer uptime: 1 Hour(s) 2 Minute(s)","version":"AzerothCore rev. fake","connected_players":3,"characters_in_world":5,"connection_peak":10,"queue":0,"uptime":"1 Hour(s) 2 Minute(s)"}` + "\n"
	if rec.Body.String() != want {
		t.Fatalf("body = %s\nwant = %s", rec.Body.String(), want)
	}
}

func TestHandleStatusMapsUpstreamFailure(t *testing.T) {
	plugin := New(fakeExecutor{err: errors.New("soap down")})

	rec := httptest.NewRecorder()
	plugin.handleStatus(rec, httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/info/status", nil))

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", rec.Code)
	}
}
