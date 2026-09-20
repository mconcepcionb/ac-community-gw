// Package domain holds the reports module's shared types.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrReportNotFound is returned when a report does not exist.
var ErrReportNotFound = errors.New("reports: report not found")

// Report statuses.
const (
	StatusOpen   = "open"
	StatusClosed = "closed"
)

// Report is a player-submitted report about another player.
type Report struct {
	ID         uuid.UUID
	ReporterID uuid.UUID
	Target     string
	Category   string
	Message    string
	Status     string
	CreatedAt  time.Time
}
