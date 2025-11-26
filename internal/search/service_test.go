package search_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/obs"
	"github.com/example/mini-hotel-aggregator/internal/search"
	"github.com/prometheus/client_golang/prometheus"
)


type mockAggregator struct {
	mu         sync.Mutex
	counter    int
	searchFunc func(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error)
}

func (m *mockAggregator) Search(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counter++

	if m.searchFunc != nil {
		return m.searchFunc(ctx, req)
	}
	return search.AggregatedResult{}, nil
}

type mockCache struct {
	getOrComputeFunc func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error)
}

func (m *mockCache) GetOrCompute(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
	if m.getOrComputeFunc != nil {
		return m.getOrComputeFunc(ctx, key, fn)
	}
	return fn(ctx)
}

func TestService_Search_Success(t *testing.T) {
	cacheCalled := false
	aggCalled := false

	cache := &mockCache{
		getOrComputeFunc: func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
			cacheCalled = true
			return fn(ctx)
		},
	}

	agg := &mockAggregator{
		searchFunc: func(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error) {
			aggCalled = true
			return search.AggregatedResult{
				Hotels: []search.Hotel{{HotelID: "H1", Name: "A", Price: 100, Nights: req.Nights}},
				Stats: struct {
					ProvidersTotal     int    "json:\"providers_total\""
					ProvidersSucceeded int    "json:\"providers_succeeded\""
					ProvidersFailed    int    "json:\"providers_failed\""
					Cache              string "json:\"cache\""
					DurationMs         int64  "json:\"duration_ms\""
				}{ProvidersTotal: 1, ProvidersSucceeded: 1, ProvidersFailed: 0, Cache: "miss"},
			}, nil
		},
	}

	svc := search.NewService(agg, cache, obs.NewMetrics(prometheus.NewRegistry()), 2*time.Second)

	req := &models.SearchRequest{City: "NYC", Checkin: "2025-11-20", Nights: 2, Adults: 2}

	res, err := svc.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cacheCalled {
		t.Fatal("expected cache to be called")
	}
	if !aggCalled {
		t.Fatal("expected aggregator to be called")
	}
	if len(res.Hotels) != 1 || res.Hotels[0].HotelID != "H1" {
		t.Fatalf("unexpected hotels: %+v", res.Hotels)
	}
}

func TestService_Search_CacheError(t *testing.T) {
	cache := &mockCache{
		getOrComputeFunc: func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
			return search.AggregatedResult{}, errors.New("cache failed")
		},
	}
	agg := &mockAggregator{}

	svc := search.NewService(agg, cache, obs.NewMetrics(prometheus.NewRegistry()), 2*time.Second)

	req := &models.SearchRequest{City: "NYC", Checkin: "2025-11-20", Nights: 1, Adults: 1}

	_, err := svc.Search(context.Background(), req)
	if err == nil || err.Error() != "cache failed" {
		t.Fatalf("expected cache error, got %v", err)
	}
}

func TestService_Search_AggregatorError(t *testing.T) {
	cache := &mockCache{
		getOrComputeFunc: func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
			// call aggregator to simulate error
			return fn(ctx)
		},
	}
	agg := &mockAggregator{
		searchFunc: func(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error) {
			return search.AggregatedResult{}, errors.New("aggregator failed")
		},
	}

	svc := search.NewService(agg, cache, obs.NewMetrics(prometheus.NewRegistry()), 2*time.Second)
	req := &models.SearchRequest{City: "NYC", Checkin: "2025-11-20", Nights: 2, Adults: 2}

	_, err := svc.Search(context.Background(), req)
	if err == nil || err.Error() != "aggregator failed" {
		t.Fatalf("expected aggregator error, got %v", err)
	}
}

func TestService_Search_Timeout(t *testing.T) {
	cache := &mockCache{
		getOrComputeFunc: func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
			time.Sleep(100 * time.Millisecond)
			return fn(ctx)
		},
	}
	agg := &mockAggregator{
		searchFunc: func(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error) {
			select {
			case <-ctx.Done():
				return search.AggregatedResult{}, ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return search.AggregatedResult{}, nil
			}
		},
	}

	svc := search.NewService(agg, cache, obs.NewMetrics(prometheus.NewRegistry()), 50*time.Millisecond)
	req := &models.SearchRequest{City: "NYC", Checkin: "2025-11-20", Nights: 2, Adults: 2}

	_, err := svc.Search(context.Background(), req)
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded error, got %v", err)
	}
}

func TestService_Search_ConcurrentRequests(t *testing.T) {
	cache := &mockCache{
		getOrComputeFunc: func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
			return fn(ctx)
		},
	}

	agg := &mockAggregator{
		searchFunc: func(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error) {
			return search.AggregatedResult{
				Hotels: []search.Hotel{{HotelID: "H1", Name: "A", Price: 50, Nights: req.Nights}},
			}, nil
		},
	}

	svc := search.NewService(agg, cache, obs.NewMetrics(prometheus.NewRegistry()), 2*time.Second)
	req := &models.SearchRequest{City: "NYC", Checkin: "2025-11-20", Nights: 1, Adults: 1}

	// simulate 5 concurrent requests
	done := make(chan struct{})
	for i := 0; i < 5; i++ {
		go func() {
			svc.Search(context.Background(), req)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 5; i++ {
		<-done
	}
	if agg.counter != 5 {
		t.Fatalf("expected aggregator to be called 5 times (no caching collapse here), got %d", agg.counter)
	}
}
