package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/pressly/goose/v3"

	"github.com/mconcepcionb/ac-community-gw/migrations"
)

// gooseMu serializes access to the goose package globals (base FS and dialect),
// which are process-wide and not safe for concurrent use.
var gooseMu sync.Mutex

// Migrate applies Goose migrations embedded in the migrations package.
//
// command is a Goose command such as "up", "down", "status" or "version".
func Migrate(ctx context.Context, db *DB, command string) error {
	gooseMu.Lock()
	defer gooseMu.Unlock()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("postgres: set goose dialect: %w", err)
	}
	if err := goose.RunContext(ctx, command, db.db, "."); err != nil {
		return fmt.Errorf("postgres: goose %s: %w", command, err)
	}
	return nil
}
