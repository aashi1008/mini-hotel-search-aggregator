package search

import (
	"context"
	"testing"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/obs"
	"github.com/prometheus/client_golang/prometheus"
)

func TestCacheCollapse(t *testing.T) {
	m := obs.NewMetrics(prometheus.NewRegistry())
	cache := NewCache(2 * time.Second, m)
	calls := 0
	fn := func(ctx context.Context) (AggregatedResult, error) {
		calls++
		time.Sleep(50 * time.Millisecond)
		return AggregatedResult{}, nil
	}

	ctx := context.Background()
	// concurrent callers
	done := make(chan struct{})
	for i := 0; i < 5; i++ {
		go func() {
			cache.GetOrCompute(ctx, "k", fn)
			done <- struct{}{}
		}()
	}
	// wait
	for i := 0; i < 5; i++ {
		<-done
	}
	if calls != 1 {
		t.Fatalf("expected single compute got %d", calls)
	}
}
