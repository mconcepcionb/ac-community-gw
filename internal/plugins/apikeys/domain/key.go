// Package domain holds the apikeys module's shared types.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrKeyNotFound is returned when an API key does not exist.
var ErrKeyNotFound = errors.New("apikeys: key not found")

// APIKey is an issued service credential. The secret is never stored.
type APIKey struct {
	ID          uuid.UUID
	Name        string
	KeyPrefix   string
	Permissions []string
	CreatedAt   time.Time
	LastUsedAt  *time.Time
}
