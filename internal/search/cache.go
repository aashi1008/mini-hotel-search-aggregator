package search

import (
	"context"
	"sync"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/obs"
)

type CacheService interface {
    GetOrCompute(ctx context.Context, key string, fn func(ctx context.Context) (AggregatedResult, error)) (AggregatedResult, error)
}

type cacheEntry struct {
	val     AggregatedResult
	expiry  time.Time
	ready   bool
	waiters []chan resultOrErr
}

type resultOrErr struct {
	res AggregatedResult
	err error
}

type Cache struct {
	mu    sync.Mutex
	ttl   time.Duration
	items map[string]*cacheEntry
	metrics *obs.Metrics
}

func NewCache(ttl time.Duration, m *obs.Metrics) *Cache {
	return &Cache{ttl: ttl, items: make(map[string]*cacheEntry), metrics: m}
}

func (c *Cache) GetOrCompute(ctx context.Context, key string, fn func(ctx context.Context) (AggregatedResult, error)) (AggregatedResult, error) {
	c.mu.Lock()
	entry, found := c.items[key]
	now := time.Now()

	// If cached and fresh, return it
	if found && entry.ready && now.Before(entry.expiry) {
		val := entry.val
		c.mu.Unlock()
		if c.metrics!=nil{
			c.metrics.IncCacheHits()
		}
		return val, nil
	}

	// Collapse: if computation in progress, join waiters
	if found && !entry.ready {
		ch := make(chan resultOrErr, 1)
		entry.waiters = append(entry.waiters, ch)
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return AggregatedResult{}, ctx.Err()
		case r := <-ch:
			return r.res, r.err
		}
	}

	// Start new computation and mark as in-flight
	ch := make(chan resultOrErr, 1)
	entry = &cacheEntry{waiters: []chan resultOrErr{ch}}
	c.items[key] = entry
	c.mu.Unlock()

	// Actual computation (only one goroutine does this)
	res, err := fn(ctx)
	result := resultOrErr{res: res, err: err}

	// Save result and notify waiters
	c.mu.Lock()
	entry.val = res
	entry.expiry = now.Add(c.ttl)
	entry.ready = true
	waiters := entry.waiters
	entry.waiters = nil
	c.mu.Unlock()

	for _, w := range waiters {
		w <- result
		close(w)
	}

	return res, err
}
