package search

import (
	"sync"
	"time"
)

// Simple token bucket per IP
type ipBucket struct {
	tokens     int
	lastRefill time.Time
}

type IPRateLimiter struct {
	mu     sync.Mutex
	buckets map[string]*ipBucket
	cap    int
	refillDuration time.Duration
}

func NewIPRateLimiter(cap int, refill time.Duration) *IPRateLimiter {
	return &IPRateLimiter{buckets: make(map[string]*ipBucket), cap: cap, refillDuration: refill}
}

func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.buckets[ip]
	now := time.Now()
	if !ok {
		rl.buckets[ip] = &ipBucket{tokens: rl.cap - 1, lastRefill: now}
		return true
	}
	// refill if interval passed
	if now.Sub(b.lastRefill) >= rl.refillDuration {
		b.tokens = rl.cap
		b.lastRefill = now
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}
