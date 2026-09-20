// Package domain holds the store module's shared types.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrProductNotFound is returned when a SKU does not exist.
	ErrProductNotFound = errors.New("store: product not found")
	// ErrProductExists is returned when a SKU is already taken.
	ErrProductExists = errors.New("store: product already exists")
	// ErrInsufficientFunds is returned when the wallet cannot cover the price.
	ErrInsufficientFunds = errors.New("store: insufficient funds")
	// ErrOrderNotFound is returned when an order does not exist.
	ErrOrderNotFound = errors.New("store: order not found")
	// ErrOrderNotPending is returned when an order is already in a terminal
	// state and the requested transition would be a no-op or a double refund.
	ErrOrderNotPending = errors.New("store: order is not pending")
)

// ProductItem is one item stack granted by a product.
type ProductItem struct {
	ItemID int
	Count  int
}

// Product is a purchasable entry of the catalog.
type Product struct {
	ID          uuid.UUID
	SKU         string
	Name        string
	Description string
	PricePoints int64
	Money       int64
	Active      bool
	Items       []ProductItem
}

// OrderStatus is the lifecycle of an order.
type OrderStatus string

const (
	// OrderPending means points were taken but delivery has not completed.
	OrderPending OrderStatus = "pending"
	// OrderDelivered means the reward reached the character.
	OrderDelivered OrderStatus = "delivered"
	// OrderFailed means delivery failed and points were refunded.
	OrderFailed OrderStatus = "failed"
)

// Order is a purchase and its delivery outcome.
type Order struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ProductID     uuid.UUID
	SKU           string
	PricePoints   int64
	CharacterName string
	AccountID     int64
	Status        OrderStatus
	CommandOutput string
	CreatedAt     time.Time
}
