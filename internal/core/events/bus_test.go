package events

import (
	"context"
	"errors"
	"testing"
)

type testEvent struct {
	Value string
}

func (testEvent) EventName() string { return "test.event" }

type otherEvent struct{}

func (otherEvent) EventName() string { return "test.other" }

func TestPublishRunsHandlersInOrder(t *testing.T) {
	bus := NewBus()
	var order []string

	_ = bus.Subscribe("test.event", func(context.Context, Event) error {
		order = append(order, "first")
		return nil
	})
	_ = bus.Subscribe("test.event", func(context.Context, Event) error {
		order = append(order, "second")
		return nil
	})

	if err := bus.Publish(context.Background(), testEvent{Value: "x"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Fatalf("unexpected handler order: %v", order)
	}
}

func TestPublishJoinsErrorsAndKeepsGoing(t *testing.T) {
	bus := NewBus()
	secondRan := false

	_ = bus.Subscribe("test.event", func(context.Context, Event) error {
		return errors.New("boom")
	})
	_ = bus.Subscribe("test.event", func(context.Context, Event) error {
		secondRan = true
		return nil
	})

	err := bus.Publish(context.Background(), testEvent{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !secondRan {
		t.Fatal("second handler did not run after first failed")
	}
}

func TestSubscribeTyped(t *testing.T) {
	bus := NewBus()
	var got string

	if err := SubscribeTyped(bus, func(_ context.Context, event testEvent) error {
		got = event.Value
		return nil
	}); err != nil {
		t.Fatalf("SubscribeTyped: %v", err)
	}

	if err := bus.Publish(context.Background(), testEvent{Value: "hello"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestPublishOtherEventDoesNotTriggerHandlers(t *testing.T) {
	bus := NewBus()
	called := false
	_ = bus.Subscribe("test.event", func(context.Context, Event) error {
		called = true
		return nil
	})

	if err := bus.Publish(context.Background(), otherEvent{}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if called {
		t.Fatal("handler ran for an unrelated event")
	}
}

func TestSubscribeRejectsEmptyName(t *testing.T) {
	bus := NewBus()
	if err := bus.Subscribe("", func(context.Context, Event) error { return nil }); err == nil {
		t.Fatal("expected error")
	}
}

func TestHandlerCount(t *testing.T) {
	bus := NewBus()
	_ = bus.Subscribe("test.event", func(context.Context, Event) error { return nil })
	_ = bus.Subscribe("test.event", func(context.Context, Event) error { return nil })
	if count := bus.HandlerCount("test.event"); count != 2 {
		t.Fatalf("HandlerCount = %d", count)
	}
}
