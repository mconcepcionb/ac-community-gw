package azerothitem

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

type fakeItems struct {
	items   []azerothdb.Item
	last    azerothdb.ItemQuery
	findErr error
}

func (f *fakeItems) ListItems(_ context.Context, query azerothdb.ItemQuery) ([]azerothdb.Item, error) {
	f.last = query
	return f.items, nil
}

func (f *fakeItems) FindItem(_ context.Context, entry int64) (azerothdb.Item, error) {
	if f.findErr != nil {
		return azerothdb.Item{}, f.findErr
	}
	for _, item := range f.items {
		if item.Entry == entry {
			return item, nil
		}
	}
	return azerothdb.Item{}, azerothdb.ErrItemNotFound
}

func TestHandleListItems(t *testing.T) {
	reader := &fakeItems{items: []azerothdb.Item{{Entry: 4496, Name: "Traveler's Backpack", Class: 1, Quality: 1}}}
	plugin := New(Config{Items: reader})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/items?filter=bag&class=1&limit=10", nil)
	rec := httptest.NewRecorder()
	plugin.handleListItems(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if reader.last.Filter != "bag" || reader.last.Class != 1 || reader.last.Limit != 10 {
		t.Fatalf("query = %+v", reader.last)
	}
	if !strings.Contains(rec.Body.String(), "Backpack") || !strings.Contains(rec.Body.String(), "Container") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestHandleGetItem(t *testing.T) {
	reader := &fakeItems{items: []azerothdb.Item{{Entry: 19019, Name: "Thunderfury", Class: 2, Quality: 5, ItemLevel: 80}}}
	plugin := New(Config{Items: reader})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/items/19019", nil)
	req.SetPathValue("entry", "19019")
	rec := httptest.NewRecorder()
	plugin.handleGetItem(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Legendary") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestHandleGetItemNotFound(t *testing.T) {
	plugin := New(Config{Items: &fakeItems{}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/items/1", nil)
	req.SetPathValue("entry", "1")
	rec := httptest.NewRecorder()
	plugin.handleGetItem(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleListItemsUnavailable(t *testing.T) {
	plugin := New(Config{})
	rec := httptest.NewRecorder()
	plugin.handleListItems(rec, httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/items", nil))

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "item_db_not_configured") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListItemsResponseShape(t *testing.T) {
	reader := &fakeItems{items: []azerothdb.Item{{Entry: 4496, Name: "Traveler's Backpack", Class: 1, Quality: 1}}}
	plugin := New(Config{Items: reader})

	rec := httptest.NewRecorder()
	plugin.handleListItems(rec, httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/items", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload ItemsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0].Entry != 4496 {
		t.Fatalf("payload = %+v", payload)
	}
	if payload.Items[0].Stats == nil || payload.Items[0].Damage == nil || payload.Items[0].Spells == nil {
		t.Fatalf("array fields must be non-null: %+v", payload.Items[0])
	}
}

func TestListItemsEmptyReturnsArray(t *testing.T) {
	plugin := New(Config{Items: &fakeItems{}})

	rec := httptest.NewRecorder()
	plugin.handleListItems(rec, httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/items", nil))
	if got := rec.Body.String(); got != `{"items":[]}`+"\n" {
		t.Fatalf("body = %q", got)
	}
}
