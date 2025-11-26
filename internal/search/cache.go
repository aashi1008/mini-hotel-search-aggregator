package search

import (
	"context"
	"sync"
	"time"
)

// simple in-memory cache with TTL and request coalescing (singleflight style)

type cacheEntry struct {
	val       AggregatedResult
	expiry    time.Time
	ready     bool
	waiters   []chan resultOrErr
}

type resultOrErr struct {
	res AggregatedResult
	err error
}

type Cache struct {
	mu    sync.Mutex
	ttl   time.Duration
	items map[string]*cacheEntry
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl, items: make(map[string]*cacheEntry)}
}

// GetOrCompute returns cached result if fresh, otherwise runs fn once and collapses concurrent calls.
func (c *Cache) GetOrCompute(ctx context.Context, key string, fn func(ctx context.Context) (AggregatedResult, error)) (AggregatedResult, bool) {
	c.mu.Lock()
	if e, ok := c.items[key]; ok && time.Now().Before(e.expiry) && e.ready {
		val := e.val
		c.mu.Unlock()
		return val, true
	}
	// if an in-flight entry exists, join its waiters
	if e, ok := c.items[key]; ok && !e.ready {
		ch := make(chan resultOrErr, 1)
		e.waiters = append(e.waiters, ch)
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return AggregatedResult{}, false
		case r := <-ch:
			return r.res, r.err == nil
		}
	}
	// no entry or expired: start a new computation and mark as in-flight
	ch := make(chan resultOrErr, 1)
	entry := &cacheEntry{waiters: []chan resultOrErr{ch}}
	c.items[key] = entry
	c.mu.Unlock()

	res, err := fn(ctx)

	r := resultOrErr{res: res, err: err}

	c.mu.Lock()
	entry, _ = c.items[key]
	entry.val = res
	entry.expiry = time.Now().Add(c.ttl)
	entry.ready = true
	ws := entry.waiters
	entry.waiters = nil
	c.mu.Unlock()

	for _, w := range ws {
		select {
		case w <- r:
		default:
		}
		close(w)
	}

	return res, err == nil
}
