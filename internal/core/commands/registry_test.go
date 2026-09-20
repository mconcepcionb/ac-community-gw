package commands

import (
	"context"
	"errors"
	"testing"
)

func TestRegisterAndExecuteTyped(t *testing.T) {
	registry := NewRegistry()
	type request struct{ Value int }
	type response struct{ Doubled int }

	err := RegisterTyped(registry, "math.double", func(_ context.Context, req request) (response, error) {
		return response{Doubled: req.Value * 2}, nil
	})
	if err != nil {
		t.Fatalf("RegisterTyped: %v", err)
	}

	result, err := ExecuteTyped[request, response](context.Background(), registry, "math.double", request{Value: 21})
	if err != nil {
		t.Fatalf("ExecuteTyped: %v", err)
	}
	if result.Doubled != 42 {
		t.Fatalf("Doubled = %d", result.Doubled)
	}
}

func TestRegisterRejectsDuplicate(t *testing.T) {
	registry := NewRegistry()
	handler := func(context.Context, any) (any, error) { return nil, nil }
	if err := registry.Register("a.b", handler); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := registry.Register("a.b", handler); !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("expected ErrAlreadyRegistered, got %v", err)
	}
}

func TestExecuteNotFound(t *testing.T) {
	registry := NewRegistry()
	if _, err := registry.Execute(context.Background(), "missing", nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRegisterTypedInvalidPayload(t *testing.T) {
	registry := NewRegistry()
	type request struct{}
	_ = RegisterTyped(registry, "a.b", func(_ context.Context, _ request) (string, error) { return "", nil })

	_, err := registry.Execute(context.Background(), "a.b", "wrong")
	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("expected ErrInvalidPayload, got %v", err)
	}
}

func TestNamesAreSorted(t *testing.T) {
	registry := NewRegistry()
	handler := func(context.Context, any) (any, error) { return nil, nil }
	_ = registry.Register("c", handler)
	_ = registry.Register("a", handler)
	_ = registry.Register("b", handler)

	names := registry.Names()
	if len(names) != 3 || names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Fatalf("unexpected names: %v", names)
	}
}
