package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewSelectsJSONHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := New("debug", "json", &buf)
	logger.Debug("hello")

	out := buf.String()
	if !strings.Contains(out, `"level":"DEBUG"`) || !strings.Contains(out, `"msg":"hello"`) {
		t.Fatalf("json output = %s", out)
	}
}

func TestNewTextHandlerHonorsLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New("info", "text", &buf)
	logger.Debug("below-level")
	logger.Info("visible")

	out := buf.String()
	if strings.Contains(out, "below-level") {
		t.Fatalf("debug must be filtered at info level: %s", out)
	}
	if !strings.Contains(out, "level=INFO") || !strings.Contains(out, "msg=visible") {
		t.Fatalf("text output = %s", out)
	}
}

func TestNewDefaultsToJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := New("info", "", &buf)
	logger.Info("x")

	if !strings.HasPrefix(strings.TrimSpace(buf.String()), "{") {
		t.Fatalf("expected JSON output, got %s", buf.String())
	}
}
