// Package storeview defines read-only cross-plugin views of store data.
//
// The views carry no behaviour and depend on nothing but the standard library,
// so a plugin can publish store reads to another plugin through the service
// registry without importing the store's domain package.
package storeview

import (
	"time"

	"github.com/google/uuid"
)

// Order is a read-only view of a store order.
type Order struct {
	ID        uuid.UUID
	SKU       string
	Points    int64
	Character string
	Status    string
	CreatedAt time.Time
}
