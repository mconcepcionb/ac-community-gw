package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	fake "github.com/mconcepcionb/ac-community-gw/internal/fake/azerothcore"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fake-azerothcore:", err)
		os.Exit(1)
	}
}

func run() error {
	var addr, user, pass, seedPath, loginDBDSN, allowedOrigins string
	var journalCapacity int
	flag.StringVar(&addr, "addr", "127.0.0.1:7878", "listen address")
	flag.StringVar(&user, "user", "", "HTTP Basic Auth username (empty disables auth; requires a loopback address)")
	flag.StringVar(&pass, "pass", "", "HTTP Basic Auth password")
	flag.StringVar(&seedPath, "state", "", "optional JSON file with initial accounts")
	flag.IntVar(&journalCapacity, "journal", 200, "number of commands kept in the in-memory journal")
	flag.StringVar(&loginDBDSN, "login-db-dsn", os.Getenv("ACGW_AZEROTH_LOGIN_DB_DSN"),
		"optional AzerothCore login DB DSN to mirror account state (default $ACGW_AZEROTH_LOGIN_DB_DSN)")
	flag.StringVar(&allowedOrigins, "allowed-origins", "", "comma-separated CORS allowlist (empty = same-origin only)")
	flag.Parse()

	if user == "" {
		if err := requireLoopback(addr); err != nil {
			return err
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	var seed []fake.Account
	if seedPath != "" {
		loaded, err := fake.LoadSeed(seedPath)
		if err != nil {
			return err
		}
		seed = loaded
		logger.Info("fake-azerothcore: loaded seed", "accounts", len(seed), "path", seedPath)
	}

	server := fake.New(fake.Config{
		Username:        user,
		Password:        pass,
		Logger:          logger,
		Seed:            seed,
		JournalCapacity: journalCapacity,
		LoginDBDSN:      loginDBDSN,
		AllowedOrigins:  splitOrigins(allowedOrigins),
	})
	defer func() { _ = server.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("fake-azerothcore: listening", "addr", addr, "auth", user != "")
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		logger.Info("fake-azerothcore: shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}

// requireLoopback refuses to start the double without authentication unless it
// is bound to a loopback address.
func requireLoopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid -addr %q: %w", addr, err)
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("refusing to run without auth on non-loopback address %q; set -user and -pass", addr)
	}
	return nil
}

func splitOrigins(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
