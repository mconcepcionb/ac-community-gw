package gatewayadmin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/storeview"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
)

type fakeUserAdmin struct {
	user  userdir.User
	roles []string
}

func (f fakeUserAdmin) UserByID(context.Context, uuid.UUID) (userdir.User, error) {
	return f.user, nil
}

func (f fakeUserAdmin) Roles(context.Context, uuid.UUID) ([]string, error) { return f.roles, nil }

type fakeAccountDirectory struct {
	username  string
	accountID *int64
}

func (f fakeAccountDirectory) LinkedAccount(context.Context, string) (string, *int64, error) {
	return f.username, f.accountID, nil
}

type fakeCharacterDirectory struct {
	characters []azerothdb.Character
}

func (f fakeCharacterDirectory) CharactersByUser(context.Context, string) ([]azerothdb.Character, error) {
	return f.characters, nil
}

type fakeStoreAccount struct {
	balance int64
	orders  []storeview.Order
}

func (f fakeStoreAccount) Wallet(context.Context, uuid.UUID) (int64, error) { return f.balance, nil }

func (f fakeStoreAccount) Orders(context.Context, uuid.UUID, int, int) ([]storeview.Order, error) {
	return f.orders, nil
}

func TestHandleUser360(t *testing.T) {
	userID := uuid.New()
	accountID := int64(42)
	plugin := New(Config{})
	registry := services.NewRegistry()
	publish := func(name string, value any) {
		if err := registry.Publish(name, value); err != nil {
			t.Fatalf("publish %s: %v", name, err)
		}
	}
	publish(identityUserAdminService, fakeUserAdmin{
		user: userdir.User{
			ID:          userID,
			DiscordID:   "123",
			Username:    "alice",
			DisplayName: "Alice",
			CreatedAt:   time.Unix(0, 0).UTC(),
		},
		roles: []string{"member"},
	})
	publish(accountDirectoryService, fakeAccountDirectory{username: "ADMIN", accountID: &accountID})
	publish(characterDirectoryService, fakeCharacterDirectory{characters: []azerothdb.Character{
		{GUID: 7, Name: "Thrall", Level: 80, Race: 2, Class: 7, GuildName: "Horde", Online: true, Money: 100},
	}})
	publish(storeAccountService, fakeStoreAccount{
		balance: 250,
		orders: []storeview.Order{
			{ID: uuid.New(), SKU: "starter", Points: 100, Character: "Thrall", Status: "delivered"},
		},
	})
	plugin.registry = registry

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/"+userID.String(), nil)
	req.SetPathValue("id", userID.String())
	rec := httptest.NewRecorder()
	plugin.handleUser360(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var got AdminUser
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.UserID != userID.String() || got.Wallet != 250 {
		t.Fatalf("unexpected response: %+v", got)
	}
	if got.AccountUsername == nil || *got.AccountUsername != "ADMIN" {
		t.Fatalf("account username = %v", got.AccountUsername)
	}
	if len(got.Characters) != 1 || got.Characters[0].Name != "Thrall" {
		t.Fatalf("characters = %+v", got.Characters)
	}
	if len(got.Orders) != 1 || got.Orders[0].SKU != "starter" {
		t.Fatalf("orders = %+v", got.Orders)
	}
}

func TestHandleUser360InvalidID(t *testing.T) {
	plugin := New(Config{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/nope", nil)
	req.SetPathValue("id", "nope")
	rec := httptest.NewRecorder()
	plugin.handleUser360(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
}
