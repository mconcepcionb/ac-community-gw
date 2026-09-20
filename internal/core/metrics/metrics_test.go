package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegistryCounters(t *testing.T) {
	registry := NewRegistry()
	registry.Inc("a")
	registry.Add("a", 2)
	registry.Inc("b")

	snapshot := registry.Snapshot()
	if snapshot["a"] != 3 || snapshot["b"] != 1 {
		t.Fatalf("snapshot = %v", snapshot)
	}

	snapshot["a"] = 99
	if registry.Snapshot()["a"] != 3 {
		t.Fatal("snapshot must be a copy")
	}
}

func TestWritePrometheus(t *testing.T) {
	registry := NewRegistry()
	registry.Inc("identity_login_started_total")

	rec := httptest.NewRecorder()
	registry.WritePrometheus(rec)

	body := rec.Body.String()
	if !strings.Contains(body, "# TYPE identity_login_started_total counter") {
		t.Fatalf("body = %q", body)
	}
	if !strings.Contains(body, "identity_login_started_total 1") {
		t.Fatalf("body = %q", body)
	}
}
