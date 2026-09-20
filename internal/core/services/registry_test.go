package services

import (
	"errors"
	"testing"
)

type greeter interface {
	Greet() string
}

type greeterImpl struct{}

func (greeterImpl) Greet() string { return "hi" }

func TestPublishAndConsumeTyped(t *testing.T) {
	registry := NewRegistry()
	if err := Provide[greeter](registry, "test.greeter", greeterImpl{}); err != nil {
		t.Fatalf("Provide: %v", err)
	}

	got, err := Consume[greeter](registry, "test.greeter")
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if got.Greet() != "hi" {
		t.Fatalf("Greet = %q", got.Greet())
	}
}

func TestPublishRejectsDuplicate(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Publish("a", greeterImpl{}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if err := registry.Publish("a", greeterImpl{}); !errors.Is(err, ErrAlreadyPublished) {
		t.Fatalf("expected ErrAlreadyPublished, got %v", err)
	}
}

func TestResolveNotFound(t *testing.T) {
	registry := NewRegistry()
	if _, err := registry.Resolve("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestConsumeTypeMismatch(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Publish("a", 42)
	if _, err := Consume[greeter](registry, "a"); err == nil {
		t.Fatal("expected type mismatch error")
	}
}
