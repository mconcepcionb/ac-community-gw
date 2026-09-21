// Package azerothmysql implements read-only access to AzerothCore's login
// database (MySQL/MariaDB).
package azerothmysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	// Register the MySQL driver with database/sql.
	_ "github.com/go-sql-driver/mysql"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

// accountColumns mirrors the information the game client can see: identity,
// email, expansion, online flag and the effective GM level across realms.
const accountColumns = `
       a.id,
       a.username,
       a.email,
       a.expansion,
       a.online,
       a.last_ip,
       a.last_login,
       COALESCE((SELECT MAX(aa.gmlevel) FROM account_access aa WHERE aa.id = a.id), 0) AS gmlevel,
       EXISTS(SELECT 1 FROM account_banned ab
              WHERE ab.id = a.id AND ab.active = 1
                AND (ab.unbandate = 0 OR ab.unbandate > UNIX_TIMESTAMP())) AS banned,
       COALESCE((SELECT ab.banreason FROM account_banned ab
                 WHERE ab.id = a.id AND ab.active = 1
                 ORDER BY ab.bandate DESC LIMIT 1), '') AS banreason`

const listAccountsQuery = `SELECT ` + accountColumns + `
FROM account a
WHERE (? = '' OR a.username LIKE ?)
ORDER BY a.username ASC
LIMIT ? OFFSET ?`

const findAccountQuery = `SELECT ` + accountColumns + `
FROM account a
WHERE a.username = ?
LIMIT 1`

// Config configures the login database client.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Client is a read-only AzerothCore login database client.
type Client struct {
	db *sql.DB
}

var _ azerothdb.AccountReader = (*Client)(nil)

// New opens and verifies a connection pool to the login database.
func New(cfg Config) (*Client, error) {
	db, err := openPool(cfg.DSN, cfg, "login")
	if err != nil {
		return nil, err
	}
	return &Client{db: db}, nil
}

// ListAccounts implements azerothdb.AccountReader.
func (c *Client) ListAccounts(ctx context.Context, query azerothdb.Query) ([]azerothdb.Account, error) {
	filter := strings.TrimSpace(query.Filter)
	limit := clampLimit(query.Limit)
	offset := clampOffset(query.Offset)
	pattern := "%" + escapeLike(filter) + "%"

	rows, err := c.db.QueryContext(ctx, listAccountsQuery, filter, pattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("azerothmysql: list accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	accounts := make([]azerothdb.Account, 0, limit)
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("azerothmysql: scan account: %w", err)
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azerothmysql: iterate accounts: %w", err)
	}
	return accounts, nil
}

// FindAccountByUsername implements azerothdb.AccountReader.
func (c *Client) FindAccountByUsername(ctx context.Context, username string) (azerothdb.Account, error) {
	row := c.db.QueryRowContext(ctx, findAccountQuery, strings.ToUpper(strings.TrimSpace(username)))
	account, err := scanAccount(row)
	if errors.Is(err, sql.ErrNoRows) {
		return azerothdb.Account{}, azerothdb.ErrAccountNotFound
	}
	if err != nil {
		return azerothdb.Account{}, fmt.Errorf("azerothmysql: find account: %w", err)
	}
	return account, nil
}

// Close closes the pool.
func (c *Client) Close() error { return c.db.Close() }

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAccount(row rowScanner) (azerothdb.Account, error) {
	var (
		account   azerothdb.Account
		lastLogin sql.NullTime
	)
	if err := row.Scan(
		&account.ID,
		&account.Username,
		&account.Email,
		&account.Expansion,
		&account.Online,
		&account.LastIP,
		&lastLogin,
		&account.GMLevel,
		&account.Banned,
		&account.BanReason,
	); err != nil {
		return azerothdb.Account{}, err
	}
	if lastLogin.Valid {
		value := lastLogin.Time
		account.LastLogin = &value
	}
	return account, nil
}
