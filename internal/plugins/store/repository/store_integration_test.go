//go:build integration

package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mconcepcionb/ac-community-gw/internal/plugins/store/domain"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("ACGW_DATABASE_URL")
	if url == "" {
		t.Skip("ACGW_DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestProductAdmin(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	store := New(db)

	sku := "itprod-" + uuid.New().String()[:8]
	created, err := store.CreateProduct(ctx, domain.Product{
		ID:          uuid.New(),
		SKU:         sku,
		Name:        "It Product",
		Description: "created",
		PricePoints: 100,
		Active:      true,
		Items:       []domain.ProductItem{{ItemID: 4496, Count: 1}, {ItemID: 6948, Count: 1}},
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM store_products WHERE id = $1`, created.ID)
	})
	if len(created.Items) != 2 {
		t.Fatalf("items = %+v", created.Items)
	}

	if _, err := store.CreateProduct(ctx, domain.Product{ID: uuid.New(), SKU: sku, Name: "dup", Active: true, Money: 1}); !errors.Is(err, domain.ErrProductExists) {
		t.Fatalf("expected ErrProductExists, got %v", err)
	}

	updated, err := store.UpdateProduct(ctx, domain.Product{
		SKU: sku, Name: "Renamed", Description: "updated", PricePoints: 250, Active: true,
		Items: []domain.ProductItem{{ItemID: 19019, Count: 1}},
	})
	if err != nil {
		t.Fatalf("UpdateProduct: %v", err)
	}
	if updated.Name != "Renamed" || updated.PricePoints != 250 || len(updated.Items) != 1 || updated.Items[0].ItemID != 19019 {
		t.Fatalf("updated = %+v", updated)
	}

	disabled, err := store.SetProductActive(ctx, sku, false)
	if err != nil {
		t.Fatalf("SetProductActive: %v", err)
	}
	if disabled.Active {
		t.Fatal("product should be inactive")
	}
}

func TestCreateProductAssignsID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	store := New(db)

	first, err := store.CreateProduct(ctx, domain.Product{
		SKU: "itid-" + uuid.New().String()[:8], Name: "A", Active: true, Money: 1,
	})
	if err != nil {
		t.Fatalf("CreateProduct first: %v", err)
	}
	second, err := store.CreateProduct(ctx, domain.Product{
		SKU: "itid-" + uuid.New().String()[:8], Name: "B", Active: true, Money: 1,
	})
	if err != nil {
		t.Fatalf("CreateProduct second: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM store_products WHERE id IN ($1, $2)`, first.ID, second.ID)
	})
	if first.ID == uuid.Nil || second.ID == uuid.Nil {
		t.Fatalf("nil id: first=%s second=%s", first.ID, second.ID)
	}
	if first.ID == second.ID {
		t.Fatalf("duplicate id: %s", first.ID)
	}
}

func TestFailOrderIsIdempotent(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	store := New(db)

	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO community_users (id, discord_id, display_name)
		VALUES ($1, $2, 'it-refund')`, userID, "itrefund-"+userID.String()); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	productID := uuid.New()
	sku := "itrefund-" + productID.String()[:8]
	if _, err := db.Exec(`INSERT INTO store_products (id, sku, name, price_points, money, active)
		VALUES ($1, $2, 'Refund Product', 100, 0, true)`, productID, sku); err != nil {
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
	if _, err := store.Grant(ctx, userID, 500, "test", uuid.Nil); err != nil {
		t.Fatalf("Grant: %v", err)
	}

	order, err := store.CreateOrder(ctx, userID, product, "Thrall", 1)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if _, err := store.FailOrder(ctx, order.ID, "boom"); err != nil {
		t.Fatalf("FailOrder: %v", err)
	}
	if balance, _ := store.Wallet(ctx, userID); balance != 500 {
		t.Fatalf("balance after refund = %d, want 500", balance)
	}
	if _, err := store.FailOrder(ctx, order.ID, "boom again"); !errors.Is(err, domain.ErrOrderNotPending) {
		t.Fatalf("second FailOrder error = %v, want ErrOrderNotPending", err)
	}
	if balance, _ := store.Wallet(ctx, userID); balance != 500 {
		t.Fatalf("balance after double refund = %d, want 500", balance)
	}

	delivered, err := store.CreateOrder(ctx, userID, product, "Thrall", 1)
	if err != nil {
		t.Fatalf("CreateOrder 2: %v", err)
	}
	if err := store.CompleteOrder(ctx, delivered.ID, "sent"); err != nil {
		t.Fatalf("CompleteOrder: %v", err)
	}
	if _, err := store.FailOrder(ctx, delivered.ID, "late"); !errors.Is(err, domain.ErrOrderNotPending) {
		t.Fatalf("FailOrder after delivery error = %v, want ErrOrderNotPending", err)
	}
	if balance, _ := store.Wallet(ctx, userID); balance != 400 {
		t.Fatalf("balance after late fail = %d, want 400", balance)
	}
	if err := store.CompleteOrder(ctx, delivered.ID, "again"); err != nil {
		t.Fatalf("idempotent CompleteOrder: %v", err)
	}
}

func TestOrderStatusConstraintRejectsUnknownStatus(t *testing.T) {
	db := openTestDB(t)
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO community_users (id, discord_id, display_name)
		VALUES ($1, $2, 'it-status')`, userID, "itstatus-"+userID.String()); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	productID := uuid.New()
	sku := "itstatus-" + productID.String()[:8]
	if _, err := db.Exec(`INSERT INTO store_products (id, sku, name, price_points, money, active)
		VALUES ($1, $2, 'Status Product', 100, 0, true)`, productID, sku); err != nil {
		t.Fatalf("seed product: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM store_orders WHERE user_id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM community_users WHERE id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM store_products WHERE id = $1`, productID)
	})
	if _, err := db.Exec(`INSERT INTO store_orders (id, user_id, product_id, sku, price_points, character_name, account_id, status)
		VALUES ($1, $2, $3, $4, 100, 'Thrall', 1, 'bogus')`, uuid.New(), userID, productID, sku); err == nil {
		t.Fatal("expected status constraint violation")
	}
}

func TestDeletingUserWithFinancialHistoryIsRestricted(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	store := New(db)

	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO community_users (id, discord_id, display_name)
		VALUES ($1, $2, 'it-restrict')`, userID, "itrestrict-"+userID.String()); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := store.Grant(ctx, userID, 100, "test", uuid.Nil); err != nil {
		t.Fatalf("Grant: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM community_users WHERE id = $1`, userID); err == nil {
		t.Fatal("expected RESTRICT when deleting a user with wallet history")
	}
	_, _ = db.Exec(`DELETE FROM store_wallet_entries WHERE user_id = $1`, userID)
	_, _ = db.Exec(`DELETE FROM store_wallets WHERE user_id = $1`, userID)
	_, _ = db.Exec(`DELETE FROM community_users WHERE id = $1`, userID)
}

func TestStoreFlow(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	store := New(db)

	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO community_users (id, discord_id, display_name)
		VALUES ($1, $2, 'it-store')`, userID, "itstore-"+userID.String()); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	productID := uuid.New()
	sku := "ittest-" + productID.String()[:8]
	if _, err := db.Exec(`INSERT INTO store_products (id, sku, name, price_points, money, active)
		VALUES ($1, $2, 'It Product', 100, 0, true)`, productID, sku); err != nil {
		t.Fatalf("seed product: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO store_product_items (product_id, item_id, count) VALUES ($1, 4496, 1)`, productID); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM community_users WHERE id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM store_products WHERE id = $1`, productID)
	})

	product, err := store.ProductBySKU(ctx, sku)
	if err != nil {
		t.Fatalf("ProductBySKU: %v", err)
	}
	if len(product.Items) != 1 || product.Items[0].ItemID != 4496 {
		t.Fatalf("product = %+v", product)
	}

	if balance, err := store.Grant(ctx, userID, 500, "test", uuid.Nil); err != nil || balance != 500 {
		t.Fatalf("Grant: balance=%d err=%v", balance, err)
	}

	order, err := store.CreateOrder(ctx, userID, product, "Thrall", 1)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if order.Status != domain.OrderPending {
		t.Fatalf("status = %s", order.Status)
	}
	if balance, _ := store.Wallet(ctx, userID); balance != 400 {
		t.Fatalf("balance after order = %d", balance)
	}

	failed, err := store.FailOrder(ctx, order.ID, "boom")
	if err != nil {
		t.Fatalf("FailOrder: %v", err)
	}
	if failed.Status != domain.OrderFailed {
		t.Fatalf("status = %s", failed.Status)
	}
	if balance, _ := store.Wallet(ctx, userID); balance != 500 {
		t.Fatalf("balance after refund = %d", balance)
	}

	second, err := store.CreateOrder(ctx, userID, product, "Thrall", 1)
	if err != nil {
		t.Fatalf("CreateOrder 2: %v", err)
	}
	if err := store.CompleteOrder(ctx, second.ID, "Mail sent."); err != nil {
		t.Fatalf("CompleteOrder: %v", err)
	}
	orders, err := store.Orders(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("Orders: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("orders = %d", len(orders))
	}

	// Drain the remaining balance (400) with four 100-point orders, then the
	// next one must fail with insufficient funds.
	for i := 0; i < 4; i++ {
		if _, err := store.CreateOrder(ctx, userID, product, "Thrall", 1); err != nil {
			t.Fatalf("CreateOrder drain %d: %v", i, err)
		}
	}
	if _, err := store.CreateOrder(ctx, userID, product, "Thrall", 1); !errors.Is(err, domain.ErrInsufficientFunds) {
		t.Fatalf("expected insufficient funds, got %v", err)
	}
}
