package httpapi

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter is an in-memory token bucket keyed by client key (usually the
// client IP). It is suitable for a single gateway instance; a distributed
// limiter can replace it behind the same Middleware contract later.
type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*bucket
	rate       float64
	burst      float64
	now        func() time.Time
	idleTTL    time.Duration
	lastSweep  time.Time
	trusted    []*net.IPNet
	maxBuckets int
}

type bucket struct {
	tokens float64
	last   time.Time
}

const defaultMaxBuckets = 100_000

// NewRateLimiter creates a limiter allowing perMinute requests with the given
// burst capacity. X-Forwarded-For is not trusted.
func NewRateLimiter(perMinute, burst int) *RateLimiter {
	return NewRateLimiterWithProxies(perMinute, burst, nil)
}

// NewRateLimiterWithProxies creates a limiter that only honors X-Forwarded-For
// when the immediate peer is within one of the trusted proxy networks.
func NewRateLimiterWithProxies(perMinute, burst int, trusted []*net.IPNet) *RateLimiter {
	if perMinute <= 0 {
		perMinute = 30
	}
	if burst <= 0 {
		burst = perMinute
	}
	return &RateLimiter{
		buckets:    make(map[string]*bucket),
		rate:       float64(perMinute) / 60.0,
		burst:      float64(burst),
		now:        time.Now,
		idleTTL:    10 * time.Minute,
		trusted:    trusted,
		maxBuckets: defaultMaxBuckets,
	}
}

// Allow reports whether the key may proceed, consuming a token.
func (l *RateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		if len(l.buckets) >= l.maxBuckets {
			l.evictOldestLocked()
		}
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	} else if elapsed := now.Sub(b.last).Seconds(); elapsed > 0 {
		b.tokens = min(l.burst, b.tokens+elapsed*l.rate)
		b.last = now
	}

	l.sweepLocked(now)
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Middleware rejects requests that exceed the limit with 429 and Retry-After.
func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(l.clientIP(r)) {
			w.Header().Set("Retry-After", strconv.Itoa(l.retryAfterSeconds()))
			WriteError(w, r, ErrTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) retryAfterSeconds() int {
	if l.rate <= 0 {
		return 60
	}
	seconds := int(1/l.rate + 0.999)
	if seconds < 1 {
		return 1
	}
	return seconds
}

func (l *RateLimiter) sweepLocked(now time.Time) {
	if now.Sub(l.lastSweep) < l.idleTTL {
		return
	}
	l.lastSweep = now
	for key, b := range l.buckets {
		if now.Sub(b.last) > l.idleTTL {
			delete(l.buckets, key)
		}
	}
}

// evictOldestLocked drops the least recently used bucket when the map is at
// capacity. Called with the mutex held.
func (l *RateLimiter) evictOldestLocked() {
	var oldestKey string
	var oldest time.Time
	for key, b := range l.buckets {
		if oldestKey == "" || b.last.Before(oldest) {
			oldestKey, oldest = key, b.last
		}
	}
	if oldestKey != "" {
		delete(l.buckets, oldestKey)
	}
}

// clientIP returns the address to rate-limit on. X-Forwarded-For is only
// honored when the immediate peer is a trusted proxy; the returned address is
// the rightmost hop that is not itself a trusted proxy.
func (l *RateLimiter) clientIP(r *http.Request) string {
	remote := remoteIP(r.RemoteAddr)
	if remote == nil || !l.isTrusted(remote) {
		return remoteString(remote, r.RemoteAddr)
	}

	var hops []net.IP
	for _, header := range r.Header.Values("X-Forwarded-For") {
		for _, part := range strings.Split(header, ",") {
			if ip := net.ParseIP(strings.TrimSpace(part)); ip != nil {
				hops = append(hops, ip)
			}
		}
	}
	for i := len(hops) - 1; i >= 0; i-- {
		if !l.isTrusted(hops[i]) {
			return hops[i].String()
		}
	}
	if len(hops) > 0 {
		return hops[0].String()
	}
	return remoteString(remote, r.RemoteAddr)
}

func (l *RateLimiter) isTrusted(ip net.IP) bool {
	for _, network := range l.trusted {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func remoteIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return net.ParseIP(strings.TrimSpace(host))
}

func remoteString(ip net.IP, fallback string) string {
	if ip == nil {
		return fallback
	}
	return ip.String()
}
