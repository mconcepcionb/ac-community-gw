package fakeazerothcore

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Config configures the fake AzerothCore SOAP server.
type Config struct {
	// Username and Password enable HTTP Basic Auth when Username is not empty.
	Username string
	Password string
	Logger   *slog.Logger
	Now      func() time.Time
	// Seed is the initial account state; it is also restored by POST /reset.
	Seed []Account
	// JournalCapacity bounds the command log kept in memory (default 200).
	JournalCapacity int
	// LoginDBDSN, when set, mirrors account state into an AzerothCore login
	// database (MySQL/MariaDB) so the gateway's read adapter sees the changes.
	LoginDBDSN string
	// AllowedOrigins is the CORS allowlist. When empty, no CORS headers are set
	// (same-origin only).
	AllowedOrigins []string
}

// Server is an in-memory AzerothCore double.
type Server struct {
	username       string
	password       string
	logger         *slog.Logger
	now            func() time.Time
	started        time.Time
	seed           []Account
	journal        *journal
	mysql          *mysqlMirror
	allowedOrigins map[string]struct{}

	mu       sync.Mutex
	accounts map[string]*Account
}

// New creates the fake server.
func New(cfg Config) *Server {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	server := &Server{
		username:       cfg.Username,
		password:       cfg.Password,
		logger:         logger,
		now:            now,
		started:        now(),
		seed:           normalizeSeed(cfg.Seed),
		journal:        newJournal(cfg.JournalCapacity),
		accounts:       make(map[string]*Account),
		allowedOrigins: newOriginSet(cfg.AllowedOrigins),
	}
	server.loadAccounts()
	mirror, err := newMySQLMirror(cfg.LoginDBDSN, logger)
	if err != nil {
		logger.Warn("fake-azerothcore: login database unavailable; running in-memory only", "error", err)
		mirror = nil
	}
	server.mysql = mirror
	if mirror != nil {
		for _, account := range server.seed {
			mirror.upsertAccount(account.Username, account.Email)
			if account.GMLevel > 0 {
				mirror.setGMLevel(account.Username, account.GMLevel)
			}
		}
	}
	return server
}

// Close releases the optional login database connection.
func (s *Server) Close() error {
	return s.mysql.close()
}

// Handler returns the SOAP endpoint, the observability API and the dashboard.
// Every route except /healthz requires the configured Basic Auth credentials.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /state", s.handleState)
	mux.HandleFunc("GET /commands", s.handleCommands)
	mux.HandleFunc("GET /commands/stream", s.handleStream)
	mux.HandleFunc("GET /", s.handleDashboard)
	mux.HandleFunc("POST /reset", s.handleReset)
	mux.HandleFunc("POST /", s.handleSOAP)
	return s.withAuth(s.withCORS(mux))
}

// withAuth enforces HTTP Basic Auth on every route except the health check and
// CORS preflight requests.
func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !s.authorized(r) {
			s.logger.Warn("fake-azerothcore: unauthorized request",
				"remote", r.RemoteAddr, "path", r.URL.Path)
			w.Header().Set("WWW-Authenticate", `Basic realm="fake-azerothcore"`)
			http.Error(w, msgUnauthorized, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withCORS echoes an allowed origin only. With no configured origins it sets no
// CORS headers, so the double is same-origin by default.
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" && s.originAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newOriginSet(origins []string) map[string]struct{} {
	set := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			set[trimmed] = struct{}{}
		}
	}
	return set
}

func (s *Server) originAllowed(origin string) bool {
	_, ok := s.allowedOrigins[origin]
	return ok
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}

func (s *Server) handleState(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.Snapshot())
}

type soapEnvelope struct {
	Body struct {
		ExecuteCommand struct {
			CommandString string `xml:"commandString"`
		} `xml:"executeCommand"`
	} `xml:"Body"`
}

func (s *Server) handleSOAP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxCommandBytes))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	var envelope soapEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		s.logger.Warn("fake-azerothcore: malformed soap request", "error", err)
		http.Error(w, "malformed soap", http.StatusBadRequest)
		return
	}
	command := strings.TrimSpace(envelope.Body.ExecuteCommand.CommandString)
	if command == "" {
		http.Error(w, "missing command", http.StatusBadRequest)
		return
	}

	requestID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
	user, _, _ := r.BasicAuth()
	start := s.now()
	result := s.execute(command)
	redacted := redactCommand(command)
	entry := s.journal.append(CommandEntry{
		At:         start,
		RequestID:  requestID,
		Command:    redacted,
		Result:     result,
		User:       user,
		Remote:     r.RemoteAddr,
		DurationMS: s.now().Sub(start).Milliseconds(),
	})

	s.logger.Info("fake-azerothcore: command",
		"seq", entry.Seq,
		"request_id", requestID,
		"command", redacted,
	)
	s.logger.Debug("fake-azerothcore: result", "seq", entry.Seq, "result", result)

	writeSOAPResult(w, result)
}

// redactCommand removes password arguments before a command is journaled or
// logged. It keeps the verb and the non-secret arguments.
func redactCommand(command string) string {
	fields := strings.Fields(command)
	if len(fields) < 2 {
		return command
	}
	redact := func(index int) {
		if index < len(fields) {
			fields[index] = "***"
		}
	}
	switch {
	case strings.EqualFold(fields[0], ".account") && len(fields) >= 2 && strings.EqualFold(fields[1], "create"):
		// .account create <name> <password> [email]
		redact(3)
	case strings.EqualFold(fields[0], ".account") && len(fields) >= 3 && strings.EqualFold(fields[1], "set") && strings.EqualFold(fields[2], "password"):
		// .account set password <name> <password> <confirmation>
		redact(4)
		redact(5)
	default:
		return command
	}
	return strings.Join(fields, " ")
}

func (s *Server) authorized(r *http.Request) bool {
	if s.username == "" {
		return true
	}
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}
	userOK := subtle.ConstantTimeCompare([]byte(user), []byte(s.username)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(pass), []byte(s.password)) == 1
	return userOK && passOK
}

func (s *Server) loadAccounts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts = make(map[string]*Account, len(s.seed))
	for _, account := range s.seed {
		copied := account
		s.accounts[normalizeAccountKey(copied.Username)] = &copied
	}
}

func normalizeSeed(seed []Account) []Account {
	out := make([]Account, 0, len(seed))
	for _, account := range seed {
		account.Username = normalizeAccountKey(account.Username)
		out = append(out, account)
	}
	return out
}

func writeSOAPResult(w http.ResponseWriter, result string) {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="urn:AC"><SOAP-ENV:Body><ns1:executeCommandResponse><result>`)
	_ = xml.EscapeText(&buf, []byte(result))
	buf.WriteString(`</result></ns1:executeCommandResponse></SOAP-ENV:Body></SOAP-ENV:Envelope>`)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}
