package azerothstore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothstore/domain"
)

type fakeCatalog struct {
	items map[int64]azerothdb.Item
}

func (f *fakeCatalog) LookupItem(_ context.Context, entry int64) (azerothdb.Item, bool, error) {
	item, ok := f.items[entry]
	return item, ok, nil
}

func TestHandleCreateProduct(t *testing.T) {
	store := newFakeStore()
	plugin := New(Config{Store: store})
	plugin.catalog = &fakeCatalog{items: map[int64]azerothdb.Item{
		4496: {Entry: 4496, Name: "Traveler's Backpack", Class: 1, Subclass: 0, Quality: 1, Description: "A 16-slot bag."},
	}}

	rec := httptest.NewRecorder()
	plugin.handleCreateProduct(rec, httptest.NewRequest(http.MethodPost, "/api/v1/store/products",
		strings.NewReader(`{"price_points":100,"items":[{"item_id":4496,"count":1}]}`)))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	created, ok := store.products["item-4496"]
	if !ok {
		t.Fatalf("product not stored: %+v", store.products)
	}
	if created.Name != "Traveler's Backpack" || created.Description != "A 16-slot bag." {
		t.Fatalf("product = %+v", created)
	}
	if !strings.Contains(rec.Body.String(), "quality_color") {
		t.Fatalf("body not enriched: %s", rec.Body.String())
	}
}

func TestHandleCreateProductUnknownItem(t *testing.T) {
	store := newFakeStore()
	plugin := New(Config{Store: store})
	plugin.catalog = &fakeCatalog{items: map[int64]azerothdb.Item{}}

	rec := httptest.NewRecorder()
	plugin.handleCreateProduct(rec, httptest.NewRequest(http.MethodPost, "/api/v1/store/products",
		strings.NewReader(`{"price_points":100,"items":[{"item_id":999,"count":1}]}`)))

	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "unknown_item") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleCreateProductBundleNeedsSKU(t *testing.T) {
	store := newFakeStore()
	plugin := New(Config{Store: store})
	plugin.catalog = &fakeCatalog{items: map[int64]azerothdb.Item{
		1: {Entry: 1, Name: "A"},
		2: {Entry: 2, Name: "B"},
	}}

	rec := httptest.NewRecorder()
	plugin.handleCreateProduct(rec, httptest.NewRequest(http.MethodPost, "/api/v1/store/products",
		strings.NewReader(`{"price_points":100,"items":[{"item_id":1,"count":1},{"item_id":2,"count":1}]}`)))

	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "missing_sku") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleDeleteProduct(t *testing.T) {
	store := newFakeStore()
	store.products["item-4496"] = domain.Product{SKU: "item-4496", Name: "Bag", Active: true}
	plugin := New(Config{Store: store})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/store/products/item-4496", nil)
	req.SetPathValue("sku", "item-4496")
	rec := httptest.NewRecorder()
	plugin.handleDeleteProduct(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.products["item-4496"].Active {
		t.Fatal("product should be inactive")
	}
}
