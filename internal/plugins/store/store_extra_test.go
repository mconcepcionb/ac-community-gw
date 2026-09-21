package store

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/store/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport"
)

// raceFailStore overrides FailOrder so the pending-state race returns
// ErrOrderNotPending.
type raceFailStore struct {
	*fakeStore
	failErr error
}

func (r *raceFailStore) FailOrder(context.Context, uuid.UUID, string) (domain.Order, error) {
	return domain.Order{}, r.failErr
}

type errCatalog struct{}

func (errCatalog) LookupItem(context.Context, int64) (azerothdb.Item, bool, error) {
	return azerothdb.Item{}, false, errors.New("catalog down")
}

func pendingOrder(store *fakeStore, sku string) domain.Order {
	order := domain.Order{
		ID: uuid.New(), SKU: sku, Status: domain.OrderPending,
		PricePoints: 100, CreatedAt: time.Now(),
	}
	store.orders[order.ID] = order
	return order
}

func TestHandleRefundOrder(t *testing.T) {
	store := newFakeStore()
	order := pendingOrder(store, "bag")
	plugin := New(Config{Store: store})

	req := testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodPost, "/x", `{"reason":"duplicate"}`), "id", order.ID.String())
	rec := testsupport.NewRecorder()
	plugin.handleRefundOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusOK)

	refunded := store.orders[order.ID]
	if refunded.Status != domain.OrderFailed || !strings.Contains(refunded.CommandOutput, "duplicate") {
		t.Fatalf("refunded = %+v", refunded)
	}
	if store.balance != 100 {
		t.Fatalf("balance = %d, want the points refunded", store.balance)
	}
}

func TestHandleRefundOrderDefaultsReason(t *testing.T) {
	store := newFakeStore()
	order := pendingOrder(store, "bag")
	plugin := New(Config{Store: store})

	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", order.ID.String())
	rec := testsupport.NewRecorder()
	plugin.handleRefundOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusOK)

	if !strings.Contains(store.orders[order.ID].CommandOutput, "manual refund: manual refund") {
		t.Fatalf("output = %q", store.orders[order.ID].CommandOutput)
	}
}

func TestHandleRefundOrderErrors(t *testing.T) {
	// nil store
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleRefundOrder(rec, testsupport.NewRequest(http.MethodPost, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	// invalid id
	store := newFakeStore()
	plugin = New(Config{Store: store})
	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", "bad")
	rec = testsupport.NewRecorder()
	plugin.handleRefundOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// order not found
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", uuid.New().String())
	rec = testsupport.NewRecorder()
	plugin.handleRefundOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)

	// not pending
	order := pendingOrder(store, "bag")
	order.Status = domain.OrderDelivered
	store.orders[order.ID] = order
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", order.ID.String())
	rec = testsupport.NewRecorder()
	plugin.handleRefundOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusConflict)

	// state race on FailOrder
	race := &raceFailStore{fakeStore: newFakeStore(), failErr: domain.ErrOrderNotPending}
	raceOrder := pendingOrder(race.fakeStore, "bag")
	plugin = New(Config{Store: race})
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", raceOrder.ID.String())
	rec = testsupport.NewRecorder()
	plugin.handleRefundOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusConflict)
}

func TestHandleRetryOrderSuccess(t *testing.T) {
	store := newFakeStore()
	order := pendingOrder(store, "bag")
	store.hasProduct = true
	store.product = domain.Product{SKU: "bag", Name: "Bag", Items: []domain.ProductItem{{ItemID: 4496, Count: 1}}}
	plugin := New(Config{Store: store})
	plugin.delivery = &fakeDelivery{}

	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", order.ID.String())
	rec := testsupport.NewRecorder()
	plugin.handleRetryOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusOK)

	if store.orders[order.ID].Status != domain.OrderDelivered {
		t.Fatalf("order = %+v", store.orders[order.ID])
	}
}

func TestHandleRetryOrderErrors(t *testing.T) {
	// nil store or delivery
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleRetryOrder(rec, testsupport.NewRequest(http.MethodPost, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	// invalid id
	store := newFakeStore()
	plugin = New(Config{Store: store})
	plugin.delivery = &fakeDelivery{}
	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", "bad")
	rec = testsupport.NewRecorder()
	plugin.handleRetryOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// order not found
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", uuid.New().String())
	rec = testsupport.NewRecorder()
	plugin.handleRetryOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)

	// not pending
	order := pendingOrder(store, "bag")
	order.Status = domain.OrderFailed
	store.orders[order.ID] = order
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", order.ID.String())
	rec = testsupport.NewRecorder()
	plugin.handleRetryOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusConflict)

	// product missing
	store.orders[order.ID] = domain.Order{ID: order.ID, SKU: "gone", Status: domain.OrderPending}
	store.hasProduct = false
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", order.ID.String())
	rec = testsupport.NewRecorder()
	plugin.handleRetryOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)

	// delivery error
	store.hasProduct = true
	store.product = domain.Product{SKU: "gone", Name: "Gone"}
	plugin.delivery = &fakeDelivery{err: delivery.ErrNotOwner}
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", order.ID.String())
	rec = testsupport.NewRecorder()
	plugin.handleRetryOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusForbidden)

	// completion failure persists the output
	plugin.delivery = &fakeDelivery{}
	store.completeErr = errors.New("db down")
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPost, "/x", ""), "id", order.ID.String())
	rec = testsupport.NewRecorder()
	plugin.handleRetryOrder(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	if store.outputSet != order.ID {
		t.Fatal("delivery output was not persisted")
	}
}

func TestHandleProductAndDeleteErrors(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleProduct(rec, testsupport.NewRequest(http.MethodGet, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{SKU: "bag", Name: "Bag"}
	plugin = New(Config{Store: store})

	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodGet, "/x", ""), "sku", "bag")
	rec = testsupport.NewRecorder()
	plugin.handleProduct(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusOK)

	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodGet, "/x", ""), "sku", "nope")
	rec = testsupport.NewRecorder()
	plugin.handleProduct(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)

	store.deactivateErr = domain.ErrProductNotFound
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodDelete, "/x", ""), "sku", "bag")
	rec = testsupport.NewRecorder()
	plugin.handleDeleteProduct(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)
}

func TestHandleOrders(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleOrders(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	store := newFakeStore()
	order := pendingOrder(store, "bag")
	plugin = New(Config{Store: store})
	rec = testsupport.NewRecorder()
	plugin.handleOrders(rec, testsupport.WithPrincipal(
		testsupport.NewRequest(http.MethodGet, "/api/v1/store/orders?limit=10", ""), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), order.SKU) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestHandleGrantValidationAndSuccess(t *testing.T) {
	// invalid points
	store := newFakeStore()
	plugin := New(Config{Store: store})
	rec := testsupport.NewRecorder()
	plugin.handleGrant(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"user_id":"`+uuid.New().String()+`","points":0}`))
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// invalid user id
	rec = testsupport.NewRecorder()
	plugin.handleGrant(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"user_id":"not-a-uuid","points":10}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	// missing user
	rec = testsupport.NewRecorder()
	plugin.handleGrant(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"points":10}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	// discord id not found
	plugin.users = &fakeUsers{found: false}
	rec = testsupport.NewRecorder()
	plugin.handleGrant(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"discord_id":"7","points":10}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	// success by user id
	userID := uuid.New()
	rec = testsupport.NewRecorder()
	plugin.handleGrant(rec, testsupport.WithPrincipal(
		testsupport.NewRequest(http.MethodPost, "/x", `{"user_id":"`+userID.String()+`","points":25}`), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusOK)
	if store.balance != 25 {
		t.Fatalf("balance = %d", store.balance)
	}

	// nil store
	rec = testsupport.NewRecorder()
	New(Config{}).handleGrant(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"points":1}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleCreateAndUpdateProductErrors(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleCreateProduct(rec, testsupport.NewRequest(http.MethodPost, "/x", `{}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	store := newFakeStore()
	plugin = New(Config{Store: store})

	// empty product
	rec = testsupport.NewRecorder()
	plugin.handleCreateProduct(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"price_points":0,"money":0}`))
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// invalid item
	rec = testsupport.NewRecorder()
	plugin.handleCreateProduct(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"price_points":1,"items":[{"item_id":0,"count":0}]}`))
	testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	// duplicate SKU
	store.createErr = domain.ErrProductExists
	rec = testsupport.NewRecorder()
	plugin.handleCreateProduct(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"sku":"bag","money":10}`))
	testsupport.AssertStatus(t, rec, http.StatusConflict)
	store.createErr = nil

	// catalog failure
	plugin.catalog = errCatalog{}
	rec = testsupport.NewRecorder()
	plugin.handleCreateProduct(rec, testsupport.NewRequest(http.MethodPost, "/x", `{"sku":"bag","price_points":1,"items":[{"item_id":1,"count":1}]}`))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	// update success (SKU from path)
	plugin.catalog = nil
	store.products["bag"] = domain.Product{SKU: "bag"}
	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"name":"New","price_points":10,"money":5}`), "sku", "bag")
	rec = testsupport.NewRecorder()
	plugin.handleUpdateProduct(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusOK)

	// update not found
	store.updateErr = domain.ErrProductNotFound
	req = testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", `{"name":"New","price_points":10,"money":5}`), "sku", "bag")
	rec = testsupport.NewRecorder()
	plugin.handleUpdateProduct(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)
}

func TestAccountView(t *testing.T) {
	plugin := New(Config{})
	if _, err := plugin.Wallet(context.Background(), uuid.New()); err == nil {
		t.Fatal("Wallet without a store must fail")
	}
	if _, err := plugin.Orders(context.Background(), uuid.New(), 10, 0); err == nil {
		t.Fatal("Orders without a store must fail")
	}

	store := newFakeStore()
	store.balance = 300
	order := pendingOrder(store, "bag")
	plugin = New(Config{Store: store})

	balance, err := plugin.Wallet(context.Background(), uuid.New())
	if err != nil || balance != 300 {
		t.Fatalf("Wallet = %d,%v", balance, err)
	}

	orders, err := plugin.Orders(context.Background(), uuid.New(), 10, 0)
	if err != nil || len(orders) != 1 {
		t.Fatalf("Orders = %+v,%v", orders, err)
	}
	if orders[0].ID != order.ID || orders[0].Points != order.PricePoints || orders[0].Status != string(domain.OrderPending) {
		t.Fatalf("order view = %+v", orders[0])
	}
}

func TestPermissionDefsAreNamespaced(t *testing.T) {
	defs := permissionDefs()
	if len(defs) != 8 {
		t.Fatalf("defs = %d, want 8", len(defs))
	}
	for _, def := range defs {
		if def.Owner != Name || def.Namespace != "gw" || !strings.HasPrefix(string(def.Name), "gw.store.") {
			t.Fatalf("definition not namespaced: %+v", def)
		}
	}
}

func newStoreTestRegistry() *plugins.Registry {
	return &plugins.Registry{
		Mux:         http.NewServeMux(),
		Services:    services.NewRegistry(),
		Permissions: permissions.NewRegistry(),
		Audit:       audit.NopRecorder{},
		RequireAuth: func(next http.Handler) http.Handler { return next },
		RequirePermission: func(_ permissions.Permission, next http.Handler) http.Handler {
			return next
		},
		RateLimit: func(next http.Handler) http.Handler { return next },
	}
}

func TestRegisterRequiresCapabilities(t *testing.T) {
	reg := newStoreTestRegistry()
	if err := New(Config{Store: newFakeStore()}).Register(context.Background(), reg); err == nil {
		t.Fatal("Register must fail when a capability is missing")
	}
}

func TestRegisterWiresStore(t *testing.T) {
	reg := newStoreTestRegistry()
	accountID := int64(5)
	if err := services.Provide[accountDirectory](reg.Services, accountDirectoryService, &fakeAccounts{linked: true, accountID: &accountID}); err != nil {
		t.Fatalf("provide accounts: %v", err)
	}
	if err := services.Provide[delivery.Service](reg.Services, deliveryService, &fakeDelivery{}); err != nil {
		t.Fatalf("provide delivery: %v", err)
	}
	if err := services.Provide[userdir.Directory](reg.Services, identityUserDirectory, &fakeUsers{}); err != nil {
		t.Fatalf("provide users: %v", err)
	}
	if err := services.Provide[azerothdb.Catalog](reg.Services, itemCatalogService, &fakeCatalog{items: map[int64]azerothdb.Item{}}); err != nil {
		t.Fatalf("provide catalog: %v", err)
	}

	store := newFakeStore()
	store.hasProduct = true
	store.product = domain.Product{SKU: "bag", Name: "Bag", Active: true}
	plugin := New(Config{Store: store})
	if err := plugin.Register(context.Background(), reg); err != nil {
		t.Fatalf("Register = %v", err)
	}

	for _, perm := range []permissions.Permission{PermissionCatalogRead, PermissionPurchase, PermissionAdminOrdersResolve} {
		if !reg.Permissions.Has(perm) {
			t.Fatalf("permission %s not registered", perm)
		}
	}
	if !reg.Services.Has(AccountService) {
		t.Fatal("AccountView capability was not published")
	}

	rec := testsupport.NewRecorder()
	reg.Mux.ServeHTTP(rec, testsupport.NewRequest(http.MethodGet, "/api/v1/store/products", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)
}
