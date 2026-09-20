// Package migrations embeds the Goose SQL migrations so they can be applied by
// the server and by the goose CLI from a single source of truth.
package migrations

import "embed"

// FS contains every SQL migration in this directory.
//
//go:embed *.sql
var FS embed.FS
