package plugins

import (
	"context"
	"errors"
	"testing"
)

type fakePlugin struct {
	name   string
	called bool
	err    error
}

func (f *fakePlugin) Name() string { return f.name }

func (f *fakePlugin) Register(context.Context, *Registry) error {
	f.called = true
	return f.err
}

func TestAddAndNamesPreserveOrder(t *testing.T) {
	manager := NewManager()
	manager.Add(&fakePlugin{name: "b"})
	manager.Add(&fakePlugin{name: "a"})

	names := manager.Names()
	if len(names) != 2 || names[0] != "b" || names[1] != "a" {
		t.Fatalf("Names = %v", names)
	}
}

func TestRegisterAllCallsEveryPlugin(t *testing.T) {
	first := &fakePlugin{name: "first"}
	second := &fakePlugin{name: "second"}
	manager := NewManager()
	manager.Add(first)
	manager.Add(second)

	if err := manager.RegisterAll(context.Background(), &Registry{}); err != nil {
		t.Fatalf("RegisterAll = %v", err)
	}
	if !first.called || !second.called {
		t.Fatalf("called = %v/%v", first.called, second.called)
	}
}

func TestRegisterAllRejectsEmptyName(t *testing.T) {
	manager := NewManager()
	manager.Add(&fakePlugin{name: ""})

	err := manager.RegisterAll(context.Background(), &Registry{})
	if !errors.Is(err, ErrEmptyName) {
		t.Fatalf("RegisterAll = %v, want ErrEmptyName", err)
	}
}

func TestRegisterAllRejectsDuplicateName(t *testing.T) {
	manager := NewManager()
	manager.Add(&fakePlugin{name: "dup"})
	manager.Add(&fakePlugin{name: "dup"})

	err := manager.RegisterAll(context.Background(), &Registry{})
	if !errors.Is(err, ErrDuplicateName) {
		t.Fatalf("RegisterAll = %v, want ErrDuplicateName", err)
	}
}

func TestRegisterAllPropagatesAndStops(t *testing.T) {
	sentinel := errors.New("register failed")
	failing := &fakePlugin{name: "failing", err: sentinel}
	after := &fakePlugin{name: "after"}
	manager := NewManager()
	manager.Add(failing)
	manager.Add(after)

	err := manager.RegisterAll(context.Background(), &Registry{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("RegisterAll = %v, want %v", err, sentinel)
	}
	if after.called {
		t.Fatal("registration must stop at the first error")
	}
}
