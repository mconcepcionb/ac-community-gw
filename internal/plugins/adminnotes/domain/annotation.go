// Package domain holds the admin notes module's shared types.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrAnnotationNotFound is returned when an annotation does not exist.
var ErrAnnotationNotFound = errors.New("adminnotes: annotation not found")

// Supported annotation target types.
const (
	TargetAccount   = "account"
	TargetUser      = "user"
	TargetCharacter = "character"
)

// Annotation is a staff note attached to a gateway entity.
type Annotation struct {
	ID         uuid.UUID
	TargetType string
	TargetID   string
	AuthorID   uuid.UUID
	Body       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ValidTargetType reports whether targetType is a supported annotation target.
func ValidTargetType(targetType string) bool {
	switch targetType {
	case TargetAccount, TargetUser, TargetCharacter:
		return true
	default:
		return false
	}
}
