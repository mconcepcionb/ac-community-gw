package ttlcache

import (
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	cache := New[string](time.Minute)
	cache.Set("k", "v")
	got, ok := cache.Get("k")
	if !ok || got != "v" {
		t.Fatalf("Get = %q,%v, want v,true", got, ok)
	}
}

func TestGetMissing(t *testing.T) {
	cache := New[int](time.Minute)
	if got, ok := cache.Get("nope"); ok || got != 0 {
		t.Fatalf("Get missing = %d,%v, want 0,false", got, ok)
	}
}

func TestExpiryEvictsTheEntry(t *testing.T) {
	cache := New[int](-time.Second)
	cache.Set("k", 1)
	if _, ok := cache.Get("k"); ok {
		t.Fatal("an entry with a past expiry must not be returned")
	}
	if _, ok := cache.Get("k"); ok {
		t.Fatal("an expired entry must stay evicted")
	}
}

func TestSetOverwrites(t *testing.T) {
	cache := New[int](time.Minute)
	cache.Set("k", 1)
	cache.Set("k", 2)
	if got, _ := cache.Get("k"); got != 2 {
		t.Fatalf("Get = %d, want 2", got)
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := New[int](time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cache.Set("k", n)
			_, _ = cache.Get("k")
		}(i)
	}
	wg.Wait()
}
