//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/store/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestOrderReadsAndReconcile(t *testing.T) {
	db := integration.Postgres(t)
	ctx := context.Background()
	store := New(db)

	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO community_users (id, discord_id, display_name)
		VALUES ($1, $2, 'it-reads')`, userID, "itreads-"+userID.String()); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	productID := uuid.New()
	sku := "itreads-" + productID.String()[:8]
	if _, err := db.Exec(`INSERT INTO store_products (id, sku, name, price_points, money, active)
		VALUES ($1, $2, 'Reads', 100, 0, true)`, productID, sku); err != nil {
		t.Fatalf("seed product: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM store_orders WHERE user_id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM store_wallet_entries WHERE user_id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM store_wallets WHERE user_id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM community_users WHERE id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM store_products WHERE id = $1`, productID)
	})

	product, err := store.ProductBySKU(ctx, sku)
	if err != nil {
		t.Fatalf("ProductBySKU: %v", err)
	}
	products, err := store.Products(ctx)
	if err != nil || !containsProduct(products, productID) {
		t.Fatalf("Products = %d, %v", len(products), err)
	}
	if _, err := store.Grant(ctx, userID, 500, "test", uuid.Nil); err != nil {
		t.Fatalf("Grant: %v", err)
	}
	order, err := store.CreateOrder(ctx, userID, product, "Thrall", 1)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if err := store.SetOrderOutput(ctx, order.ID, "Mail sent."); err != nil {
		t.Fatalf("SetOrderOutput: %v", err)
	}
	got, err := store.OrderByID(ctx, order.ID)
	if err != nil || got.Status != domain.OrderPending || got.CommandOutput != "Mail sent." {
		t.Fatalf("OrderByID = %+v, %v", got, err)
	}
	admin, err := store.AdminOrders(ctx, string(domain.OrderPending), 100, 0)
	if err != nil || !containsOrder(admin, order.ID) {
		t.Fatalf("AdminOrders = %d, %v", len(admin), err)
	}

	if _, err := store.ReconcilePendingOrders(ctx, 50); err != nil {
		t.Fatalf("ReconcilePendingOrders: %v", err)
	}
	reconciled, err := store.OrderByID(ctx, order.ID)
	if err != nil || reconciled.Status != domain.OrderDelivered {
		t.Fatalf("reconciled = %+v, %v", reconciled, err)
	}
}

func containsProduct(products []domain.Product, id uuid.UUID) bool {
	for _, product := range products {
		if product.ID == id {
			return true
		}
	}
	return false
}

func containsOrder(orders []domain.Order, id uuid.UUID) bool {
	for _, order := range orders {
		if order.ID == id {
			return true
		}
	}
	return false
}
