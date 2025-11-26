package search

import (
	"context"
	"testing"
	"time"
)

func TestCacheCollapse(t *testing.T) {
	cache := NewCache(2 * time.Second)
	calls := 0
	fn := func(ctx context.Context) (AggregatedResult, error) {
		calls++
		// simulate some work
		time.Sleep(50 * time.Millisecond)
		return AggregatedResult{}, nil
	}

	ctx := context.Background()
	// concurrent callers
	done := make(chan struct{})
	for i:=0;i<5;i++ {
		go func() {
			cache.GetOrCompute(ctx, "k", fn)
			done <- struct{}{}
		}()
	}
	// wait
	for i:=0;i<5;i++ { <-done }
	if calls != 1 { t.Fatalf("expected single compute got %d", calls) }
}
