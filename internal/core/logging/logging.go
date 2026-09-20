// Package logging builds the structured logger used across the gateway.
//
// It intentionally relies only on the standard library log/slog.
package logging

import (
	"io"
	"log/slog"
	"strings"
)

// New builds a slog.Logger writing to w.
func New(level, format string, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: ParseLevel(level)}
	var handler slog.Handler
	if strings.EqualFold(format, "text") {
		handler = slog.NewTextHandler(w, opts)
	} else {
		handler = slog.NewJSONHandler(w, opts)
	}
	return slog.New(handler)
}

// ParseLevel maps a textual level to slog.Level, defaulting to info.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
