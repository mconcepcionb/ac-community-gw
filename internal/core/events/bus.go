// Package events provides the in-process, synchronous event bus.
//
// The bus announces facts that already happened. It is not a command or query
// bus and it offers best-effort delivery only for the lifetime of the process.
// Handlers run synchronously inside Publish, in registration order.
package events

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// Event is any domain fact that can be published on the bus.
type Event interface {
	EventName() string
}

// Handler consumes a single event.
type Handler func(ctx context.Context, event Event) error

// Bus is a concurrency-safe, synchronous, in-memory event bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewBus creates an empty bus.
func NewBus() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

// Subscribe registers a handler for the given event name.
func (b *Bus) Subscribe(name string, handler Handler) error {
	if name == "" {
		return errors.New("events: event name must not be empty")
	}
	if handler == nil {
		return errors.New("events: handler must not be nil")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[name] = append(b.handlers[name], handler)
	return nil
}

// SubscribeTyped registers a typed handler for event type E. It never
// dereferences a nil pointer, so pointer event types are safe.
func SubscribeTyped[E Event](b *Bus, handler func(ctx context.Context, event E) error) error {
	name, err := EventName[E]()
	if err != nil {
		return err
	}
	return b.Subscribe(name, func(ctx context.Context, event Event) error {
		typed, ok := event.(E)
		if !ok {
			return fmt.Errorf("events: handler for %q received %T", name, event)
		}
		return handler(ctx, typed)
	})
}

// EventName resolves the event name of E without dereferencing a nil pointer.
func EventName[E Event]() (string, error) {
	var zero E
	eventType := reflect.TypeOf(zero)
	if eventType == nil {
		return "", errors.New("events: event type must not be nil")
	}
	if eventType.Kind() == reflect.Pointer {
		zero = reflect.New(eventType.Elem()).Interface().(E)
	}
	return zero.EventName(), nil
}

// Publish dispatches event to every registered handler synchronously.
//
// All handlers are invoked even if one fails; their errors are joined.
// Handlers are never started in background goroutines.
func (b *Bus) Publish(ctx context.Context, event Event) error {
	if event == nil {
		return errors.New("events: event must not be nil")
	}
	name := event.EventName()

	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[name]...)
	b.mu.RUnlock()

	var errs []error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			errs = append(errs, fmt.Errorf("events: %s handler: %w", name, err))
		}
	}
	return errors.Join(errs...)
}

// HandlerCount returns the number of handlers registered for an event name.
func (b *Bus) HandlerCount(name string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[name])
}
