package search

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewIPRateLimiter(2, time.Minute)
	if !rl.Allow("1.1.1.1") { t.Fatal("expected allow") }
	if !rl.Allow("1.1.1.1") { t.Fatal("expected allow") }
	if rl.Allow("1.1.1.1") { t.Fatal("expected deny") }
}
