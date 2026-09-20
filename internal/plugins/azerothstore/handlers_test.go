package azerothstore

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothstore/domain"
)

type fakeStore struct {
	product     domain.Product
	hasProduct  bool
	balance     int64
	orders      map[uuid.UUID]domain.Order
	products    map[string]domain.Product
	completeErr error
	outputSet   uuid.UUID
}

func newFakeStore() *fakeStore {
	return &fakeStore{orders: map[uuid.UUID]domain.Order{}, products: map[string]domain.Product{}}
}

func (f *fakeStore) CreateProduct(_ context.Context, product domain.Product) (domain.Product, error) {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}
	f.products[product.SKU] = product
	return product, nil
}

func (f *fakeStore) UpdateProduct(_ context.Context, product domain.Product) (domain.Product, error) {
	if _, ok := f.products[product.SKU]; !ok {
		return domain.Product{}, domain.ErrProductNotFound
	}
	f.products[product.SKU] = product
	return product, nil
}

func (f *fakeStore) SetProductActive(_ context.Context, sku string, active bool) (domain.Product, error) {
	product, ok := f.products[sku]
	if !ok {
		return domain.Product{}, domain.ErrProductNotFound
	}
	product.Active = active
	f.products[sku] = product
	return product, nil
}

func (f *fakeStore) Products(context.Context) ([]domain.Product, error) {
	if !f.hasProduct {
		return nil, nil
	}
	return []domain.Product{f.product}, nil
}

func (f *fakeStore) ProductBySKU(_ context.Context, sku string) (domain.Product, error) {
	if !f.hasProduct || f.product.SKU != sku {
		return domain.Product{}, domain.ErrProductNotFound
	}
	return f.product, nil
}

func (f *fakeStore) Wallet(context.Context, uuid.UUID) (int64, error) { return f.balance, nil }

func (f *fakeStore) Grant(_ context.Context, _ uuid.UUID, points int64, _ string, _ uuid.UUID) (int64, error) {
	f.balance += points
	return f.balance, nil
}

func (f *fakeStore) CreateOrder(_ context.Context, userID uuid.UUID, product domain.Product, character string, accountID int64) (domain.Order, error) {
	if f.balance < product.PricePoints {
		return domain.Order{}, domain.ErrInsufficientFunds
	}
	f.balance -= product.PricePoints
	order := domain.Order{
		ID:            uuid.New(),
		UserID:        userID,
		ProductID:     product.ID,
		SKU:           product.SKU,
		PricePoints:   product.PricePoints,
		CharacterName: character,
		AccountID:     accountID,
		Status:        domain.OrderPending,
		CreatedAt:     time.Now(),
	}
	f.orders[order.ID] = order
	return order, nil
}

func (f *fakeStore) CompleteOrder(_ context.Context, orderID uuid.UUID, output string) error {
	if f.completeErr != nil {
		return f.completeErr
	}
	order := f.orders[orderID]
	order.Status = domain.OrderDelivered
	order.CommandOutput = output
	f.orders[orderID] = order
	return nil
}

func (f *fakeStore) SetOrderOutput(_ context.Context, orderID uuid.UUID, output string) error {
	order := f.orders[orderID]
	order.CommandOutput = output
	f.orders[orderID] = order
	f.outputSet = orderID
	return nil
}

func (f *fakeStore) ReconcilePendingOrders(context.Context, int) (int, error) {
	return 0, nil
}

func (f *fakeStore) FailOrder(_ context.Context, orderID uuid.UUID, output string) (domain.Order, error) {
	order := f.orders[orderID]
	order.Status = domain.OrderFailed
	order.CommandOutput = output
	f.balance += order.PricePoints
	f.orders[orderID] = order
	return order, nil
}

func (f *fakeStore) Orders(context.Context, uuid.UUID, int, int) ([]domain.Order, error) {
	out := make([]domain.Order, 0, len(f.orders))
	for _, order := range f.orders {
		out = append(out, order)
	}
	return out, nil
}

func (f *fakeStore) AdminOrders(_ context.Context, status string, _, _ int) ([]domain.Order, error) {
	out := make([]domain.Order, 0, len(f.orders))
	for _, order := range f.orders {
		if status != "" && string(order.Status) != status {
			continue
		}
		out = append(out, order)
	}
	return out, nil
}

func (f *fakeStore) OrderByID(_ context.Context, id uuid.UUID) (domain.Order, error) {
	order, ok := f.orders[id]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return order, nil
}

type fakeDelivery struct {
	err  error
	last delivery.Request
}

func (f *fakeDelivery) Deliver(_ context.Context, req delivery.Request) (string, error) {
	f.last = req
	if f.err != nil {
		return "", f.err
	}
	return "Mail sent.", nil
}

type fakeAccounts struct {
	accountID *int64
	linked    bool
}

func (f *fakeAccounts) LinkedAccount(context.Context, string) (string, *int64, error) {
	if !f.linked {
		return "", nil, nil
	}
	return "ADMIN", f.accountID, nil
}

type fakeUsers struct {
	id    uuid.UUID
	found bool
}

func (f *fakeUsers) ResolveUser(context.Context, string) (uuid.UUID, bool, error) {
	return f.id, f.found, nil
}

func (f *fakeUsers) ResolveUserByName(context.Context, string) ([]userdir.User, error) {
	return nil, nil
}

func (f *fakeUsers) ListUsers(context.Context, string, int, int) ([]userdir.User, error) {
	return nil, nil
}

func purchasePlugin(t *testing.T, store *fakeStore, accounts *fakeAccounts, del *fakeDelivery) *Plugin {
	t.Helper()
	plugin := New(Config{Store: store})
	plugin.accounts = accounts
	plugin.delivery = del
	return plugin
}

func authedRequest(method, target, body string, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{UserID: userID, DiscordID: "42"}))
}

func TestHandleProducts(t *testing.T) {
	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{SKU: "bag-16", Name: "Bag", PricePoints: 100, Active: true}
	plugin := New(Config{Store: store})

	rec := httptest.NewRecorder()
	plugin.handleProducts(rec, httptest.NewRequest(http.MethodGet, "/api/v1/store/products", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "bag-16") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleWallet(t *testing.T) {
	store := newFakeStore()
	store.balance = 500
	plugin := New(Config{Store: store})

	rec := httptest.NewRecorder()
	plugin.handleWallet(rec, authedRequest(http.MethodGet, "/api/v1/store/wallet", "", uuid.New()))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "500") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlePurchaseSuccess(t *testing.T) {
	accountID := int64(5)
	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{ID: uuid.New(), SKU: "bag-16", Name: "Bag", PricePoints: 100, Active: true, Items: []domain.ProductItem{{ItemID: 4496, Count: 1}}}
	store.balance = 500
	del := &fakeDelivery{}
	plugin := purchasePlugin(t, store, &fakeAccounts{linked: true, accountID: &accountID}, del)

	rec := httptest.NewRecorder()
	plugin.handlePurchase(rec, authedRequest(http.MethodPost, "/api/v1/store/orders", `{"sku":"bag-16","character":"Thrall"}`, uuid.New()))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.balance != 400 {
		t.Fatalf("balance = %d", store.balance)
	}
	if del.last.AccountID != 5 || len(del.last.Items) != 1 || del.last.Items[0].ID != 4496 {
		t.Fatalf("delivery = %+v", del.last)
	}
}

func TestHandlePurchaseInsufficientFunds(t *testing.T) {
	accountID := int64(5)
	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{ID: uuid.New(), SKU: "bag-16", PricePoints: 100, Active: true}
	store.balance = 10
	plugin := purchasePlugin(t, store, &fakeAccounts{linked: true, accountID: &accountID}, &fakeDelivery{})

	rec := httptest.NewRecorder()
	plugin.handlePurchase(rec, authedRequest(http.MethodPost, "/api/v1/store/orders", `{"sku":"bag-16","character":"Thrall"}`, uuid.New()))
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlePurchaseNoLink(t *testing.T) {
	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{ID: uuid.New(), SKU: "bag-16", PricePoints: 100, Active: true}
	store.balance = 500
	plugin := purchasePlugin(t, store, &fakeAccounts{}, &fakeDelivery{})

	rec := httptest.NewRecorder()
	plugin.handlePurchase(rec, authedRequest(http.MethodPost, "/api/v1/store/orders", `{"sku":"bag-16","character":"Thrall"}`, uuid.New()))
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlePurchaseInactiveProduct(t *testing.T) {
	accountID := int64(5)
	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{ID: uuid.New(), SKU: "bag-16", PricePoints: 100, Active: false}
	store.balance = 500
	plugin := purchasePlugin(t, store, &fakeAccounts{linked: true, accountID: &accountID}, &fakeDelivery{})

	rec := httptest.NewRecorder()
	plugin.handlePurchase(rec, authedRequest(http.MethodPost, "/api/v1/store/orders", `{"sku":"bag-16","character":"Thrall"}`, uuid.New()))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.balance != 500 {
		t.Fatalf("balance changed to %d", store.balance)
	}
}

func TestHandlePurchasePersistsOutputOnCompletionFailure(t *testing.T) {
	accountID := int64(5)
	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{ID: uuid.New(), SKU: "bag-16", PricePoints: 100, Active: true}
	store.balance = 500
	store.completeErr = errors.New("db down")
	plugin := purchasePlugin(t, store, &fakeAccounts{linked: true, accountID: &accountID}, &fakeDelivery{})

	rec := httptest.NewRecorder()
	plugin.handlePurchase(rec, authedRequest(http.MethodPost, "/api/v1/store/orders", `{"sku":"bag-16","character":"Thrall"}`, uuid.New()))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.outputSet == uuid.Nil {
		t.Fatal("delivery output was not persisted for reconciliation")
	}
}

func TestHandlePurchaseDeliveryRefundsOnFailure(t *testing.T) {
	accountID := int64(5)
	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{ID: uuid.New(), SKU: "bag-16", PricePoints: 100, Active: true}
	store.balance = 500
	plugin := purchasePlugin(t, store, &fakeAccounts{linked: true, accountID: &accountID}, &fakeDelivery{err: delivery.ErrNotOwner})

	rec := httptest.NewRecorder()
	plugin.handlePurchase(rec, authedRequest(http.MethodPost, "/api/v1/store/orders", `{"sku":"bag-16","character":"Ghost"}`, uuid.New()))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.balance != 500 {
		t.Fatalf("balance not refunded: %d", store.balance)
	}
}

func TestHandleGrantByDiscordID(t *testing.T) {
	store := newFakeStore()
	plugin := New(Config{Store: store})
	userID := uuid.New()
	plugin.users = &fakeUsers{id: userID, found: true}

	rec := httptest.NewRecorder()
	plugin.handleGrant(rec, authedRequest(http.MethodPost, "/api/v1/store/wallets/grant",
		`{"discord_id":"42","points":250,"reason":"event"}`, uuid.New()))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.balance != 250 {
		t.Fatalf("balance = %d", store.balance)
	}
}

func TestHandleAdminOrders(t *testing.T) {
	store := newFakeStore()
	delivered := domain.Order{ID: uuid.New(), Status: domain.OrderDelivered, SKU: "starter"}
	pending := domain.Order{ID: uuid.New(), Status: domain.OrderPending, SKU: "mystery"}
	store.orders[delivered.ID] = delivered
	store.orders[pending.ID] = pending
	plugin := New(Config{Store: store})

	rec := httptest.NewRecorder()
	plugin.handleAdminOrders(rec, httptest.NewRequest(http.MethodGet,
		"/api/v1/admin/store/orders?status=delivered", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "starter") || strings.Contains(body, "mystery") {
		t.Fatalf("unexpected body: %s", body)
	}
}
