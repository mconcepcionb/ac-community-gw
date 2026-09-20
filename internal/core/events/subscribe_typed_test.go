package events

import (
	"context"
	"testing"
)

type valueEvent struct {
	ID string
}

func (valueEvent) EventName() string { return "value.event" }

func TestSubscribeTypedPointerEventDoesNotPanic(t *testing.T) {
	name, err := EventName[*valueEvent]()
	if err != nil {
		t.Fatalf("EventName: %v", err)
	}
	if name != "value.event" {
		t.Fatalf("name = %q", name)
	}

	bus := NewBus()
	called := false
	if err := SubscribeTyped[*valueEvent](bus, func(_ context.Context, event *valueEvent) error {
		called = true
		if event.ID != "42" {
			t.Fatalf("id = %q", event.ID)
		}
		return nil
	}); err != nil {
		t.Fatalf("SubscribeTyped: %v", err)
	}
	if err := bus.Publish(context.Background(), &valueEvent{ID: "42"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}
