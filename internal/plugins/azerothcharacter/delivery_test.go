package azerothcharacter

import (
	"context"
	"errors"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
)

func TestDeliverSendsToOwner(t *testing.T) {
	executor := &fakeExecutor{result: "Mail sent."}
	characters := &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}}
	plugin := New(Config{Executor: executor, Characters: characters})

	output, err := plugin.Deliver(context.Background(), delivery.Request{
		AccountID: 5,
		Character: "Thrall",
		Items:     []delivery.Item{{ID: 4496, Count: 1}},
	})
	if err != nil {
		t.Fatalf("Deliver: %v", err)
	}
	if executor.last != `.send items Thrall "Reward" " " 4496:1` {
		t.Fatalf("command = %q", executor.last)
	}
	if output == "" {
		t.Fatal("output is empty")
	}
}

func TestDeliverRejectsNonOwner(t *testing.T) {
	characters := &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}}
	plugin := New(Config{Executor: &fakeExecutor{}, Characters: characters})

	_, err := plugin.Deliver(context.Background(), delivery.Request{
		AccountID: 6,
		Character: "Thrall",
		Items:     []delivery.Item{{ID: 4496, Count: 1}},
	})
	if !errors.Is(err, delivery.ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner, got %v", err)
	}
}

func TestDeliverRejectsUnknownCharacter(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}, Characters: &fakeCharacters{}})
	_, err := plugin.Deliver(context.Background(), delivery.Request{
		AccountID: 5,
		Character: "Ghost",
		Money:     100,
	})
	if !errors.Is(err, delivery.ErrCharacterNotFound) {
		t.Fatalf("expected ErrCharacterNotFound, got %v", err)
	}
}
