package httpapi

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterBurstAndRefill(t *testing.T) {
	now := time.Now()
	limiter := NewRateLimiter(60, 2)
	limiter.now = func() time.Time { return now }

	if !limiter.Allow("a") {
		t.Fatal("first request of the burst should be allowed")
	}
	if !limiter.Allow("a") {
		t.Fatal("second request of the burst should be allowed")
	}
	if limiter.Allow("a") {
		t.Fatal("third request should be denied")
	}

	now = now.Add(time.Second)
	if !limiter.Allow("a") {
		t.Fatal("request should be allowed after refill")
	}
}

func TestRateLimiterIsolatesKeys(t *testing.T) {
	limiter := NewRateLimiter(60, 1)
	if !limiter.Allow("a") {
		t.Fatal("first key should be allowed")
	}
	if !limiter.Allow("b") {
		t.Fatal("second key should not be limited by the first")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	limiter := NewRateLimiter(60, 1)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "1.2.3.4:5555"
	handler.ServeHTTP(first, req)
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d", first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, req)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header missing")
	}
}

func TestClientIPIgnoresForwardedForFromUntrustedPeer(t *testing.T) {
	limiter := NewRateLimiter(60, 1)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "9.9.9.9")
	if got := limiter.clientIP(req); got != "10.0.0.1" {
		t.Fatalf("clientIP = %q, want RemoteAddr", got)
	}
}

func TestClientIPHonorsForwardedForFromTrustedPeer(t *testing.T) {
	_, network, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatalf("cidr: %v", err)
	}
	limiter := NewRateLimiterWithProxies(60, 1, []*net.IPNet{network})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "9.9.9.9, 10.0.0.1")
	if got := limiter.clientIP(req); got != "9.9.9.9" {
		t.Fatalf("clientIP = %q, want 9.9.9.9", got)
	}
}

func TestClientIPSkipsTrustedHops(t *testing.T) {
	_, network, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatalf("cidr: %v", err)
	}
	limiter := NewRateLimiterWithProxies(60, 1, []*net.IPNet{network})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "9.9.9.9, 10.0.0.5")
	if got := limiter.clientIP(req); got != "9.9.9.9" {
		t.Fatalf("clientIP = %q, want 9.9.9.9", got)
	}
}

func TestRateLimiterBucketCap(t *testing.T) {
	limiter := NewRateLimiter(60, 1)
	limiter.maxBuckets = 3
	for i := 0; i < 10; i++ {
		limiter.Allow(string(rune('a' + i)))
	}
	if len(limiter.buckets) > 3 {
		t.Fatalf("buckets = %d, want <= 3", len(limiter.buckets))
	}
}

func TestRateLimiterSpoofedForwardedForSharesBucket(t *testing.T) {
	limiter := NewRateLimiter(60, 1)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "1.2.3.4:5555"
	req.Header.Set("X-Forwarded-For", "9.9.9.9")
	handler.ServeHTTP(first, req)
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d", first.Code)
	}

	second := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req2.RemoteAddr = "1.2.3.4:5555"
	req2.Header.Set("X-Forwarded-For", "8.8.8.8")
	handler.ServeHTTP(second, req2)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want 429", second.Code)
	}
}
