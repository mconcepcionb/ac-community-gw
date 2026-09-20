package fakeazerothcore

import (
	"database/sql"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
)

// mysqlMirror optionally persists account state to an AzerothCore login
// database so the gateway's read adapter sees the same accounts the double
// serves over SOAP. It is best-effort: failures are logged, never fatal.
type mysqlMirror struct {
	db     *sql.DB
	logger *slog.Logger
}

func newMySQLMirror(dsn string, logger *slog.Logger) (*mysqlMirror, error) {
	if dsn == "" {
		return nil, nil
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &mysqlMirror{db: db, logger: logger}, nil
}

func (m *mysqlMirror) close() error {
	if m == nil || m.db == nil {
		return nil
	}
	return m.db.Close()
}

func (m *mysqlMirror) upsertAccount(username, email string) {
	if m == nil {
		return
	}
	_, err := m.db.Exec(`
		INSERT INTO account (username, salt, verifier, email, reg_mail, expansion)
		VALUES (?, UNHEX(REPEAT('00', 32)), UNHEX(REPEAT('00', 32)), ?, ?, 0)
		ON DUPLICATE KEY UPDATE email = VALUES(email)`, username, email, email)
	m.logErr("upsert account", err)
}

func (m *mysqlMirror) setEmail(username, email string) {
	if m == nil {
		return
	}
	_, err := m.db.Exec(`UPDATE account SET email = ? WHERE username = ?`, email, username)
	m.logErr("set email", err)
}

func (m *mysqlMirror) setGMLevel(username string, level int) {
	if m == nil {
		return
	}
	_, err := m.db.Exec(`
		INSERT INTO account_access (id, gmlevel, RealmID)
		SELECT id, ?, -1 FROM account WHERE username = ?
		ON DUPLICATE KEY UPDATE gmlevel = VALUES(gmlevel)`, level, username)
	m.logErr("set gm level", err)
}

func (m *mysqlMirror) setBanned(username string, seconds int, reason string) {
	if m == nil {
		return
	}
	if seconds < 0 {
		seconds = 0
	}
	// A zero unbandate means a permanent ban in AzerothCore.
	_, err := m.db.Exec(`
		INSERT INTO account_banned (id, bandate, unbandate, bannedby, banreason, active)
		SELECT id, UNIX_TIMESTAMP(),
		       CASE WHEN ? > 0 THEN UNIX_TIMESTAMP() + ? ELSE 0 END,
		       'fake-azerothcore', ?, 1
		FROM account WHERE username = ?`, seconds, seconds, reason, username)
	m.logErr("ban account", err)
}

func (m *mysqlMirror) clearBan(username string) {
	if m == nil {
		return
	}
	_, err := m.db.Exec(`
		UPDATE account_banned ab
		JOIN account a ON a.id = ab.id
		SET ab.active = 0
		WHERE a.username = ?`, username)
	m.logErr("unban account", err)
}

// deleteAccount removes an account (and its dependent rows) from the mirror.
// Used by Reset to undo accounts created after startup.
func (m *mysqlMirror) deleteAccount(username string) {
	if m == nil {
		return
	}
	if _, err := m.db.Exec(`DELETE FROM account_access WHERE id IN (SELECT id FROM account WHERE username = ?)`, username); err != nil {
		m.logErr("delete account access", err)
	}
	if _, err := m.db.Exec(`DELETE FROM account_banned WHERE id IN (SELECT id FROM account WHERE username = ?)`, username); err != nil {
		m.logErr("delete account ban", err)
	}
	if _, err := m.db.Exec(`DELETE FROM account WHERE username = ?`, username); err != nil {
		m.logErr("delete account", err)
	}
}

func (m *mysqlMirror) logErr(action string, err error) {
	if err != nil && m.logger != nil {
		m.logger.Warn("fake-azerothcore: login db "+action, "error", err)
	}
}
