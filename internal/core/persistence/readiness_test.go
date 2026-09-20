package persistence

import (
	"context"
	"errors"
	"testing"
	"time"
)

type blockingCheck struct {
	name  string
	block time.Duration
}

func (c blockingCheck) Name() string { return c.name }

func (c blockingCheck) Check(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(c.block):
		return nil
	}
}

type failingCheck struct {
	name string
	err  error
}

func (c failingCheck) Name() string                  { return c.name }
func (c failingCheck) Check(_ context.Context) error { return c.err }

func TestRegistryEvaluateTimesOutSlowCheck(t *testing.T) {
	registry := NewRegistry()
	registry.SetTimeout(20 * time.Millisecond)
	registry.Add(blockingCheck{name: "slow", block: time.Second})

	start := time.Now()
	results := registry.Evaluate(context.Background())
	elapsed := time.Since(start)

	if elapsed > 500*time.Millisecond {
		t.Fatalf("Evaluate took %v, expected the per-check timeout to fire", elapsed)
	}
	if len(results) != 1 || !errors.Is(results[0].Err, context.DeadlineExceeded) {
		t.Fatalf("results = %+v", results)
	}
}

func TestRegistryCheckJoinsErrors(t *testing.T) {
	registry := NewRegistry()
	registry.Add(failingCheck{name: "a", err: errors.New("boom")})
	registry.Add(failingCheck{name: "b", err: errors.New("bang")})
	err := registry.Check(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
