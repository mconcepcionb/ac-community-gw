package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/storeview"
)

// AccountService is the capability to read a user's wallet and orders.
const AccountService = "azeroth.store.account"

// AccountView is the cross-plugin capability published by this plugin.
type AccountView interface {
	// Wallet returns the user's point balance.
	Wallet(ctx context.Context, userID uuid.UUID) (int64, error)
	// Orders lists the user's orders, newest first.
	Orders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]storeview.Order, error)
}

// Wallet implements AccountView.
func (p *Plugin) Wallet(ctx context.Context, userID uuid.UUID) (int64, error) {
	if p.store == nil {
		return 0, errStoreUnavailable
	}
	return p.store.Wallet(ctx, userID)
}

// Orders implements AccountView.
func (p *Plugin) Orders(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]storeview.Order, error) {
	if p.store == nil {
		return nil, errStoreUnavailable
	}
	orders, err := p.store.Orders(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	views := make([]storeview.Order, 0, len(orders))
	for _, order := range orders {
		views = append(views, storeview.Order{
			ID:        order.ID,
			SKU:       order.SKU,
			Points:    order.PricePoints,
			Character: order.CharacterName,
			Status:    string(order.Status),
			CreatedAt: order.CreatedAt,
		})
	}
	return views, nil
}
