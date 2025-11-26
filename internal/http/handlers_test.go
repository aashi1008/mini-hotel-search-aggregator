package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	ht "github.com/example/mini-hotel-aggregator/internal/http"
	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/obs"
	"github.com/example/mini-hotel-aggregator/internal/search"
	"github.com/prometheus/client_golang/prometheus"
)

// ------------------------ MOCKS ------------------------
type mockAggregator struct {
	searchFunc func(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error)
}

func (m *mockAggregator) Search(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error) {
	return m.searchFunc(ctx, req)
}

type mockCache struct {
	getOrComputeFunc func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error)
}

func (m *mockCache) GetOrCompute(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
	return m.getOrComputeFunc(ctx, key, fn)
}

type mockRateLimiter struct {
	allowFunc func(ip string) bool
}

func (m *mockRateLimiter) Allow(ip string) bool {
	return m.allowFunc(ip)
}

// -------------------------------------------------------

func TestHandler_Search_Positive(t *testing.T) {
	cache := &mockCache{
		getOrComputeFunc: func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
			return search.AggregatedResult{
				Hotels: []search.Hotel{
					{HotelID: "H1", Name: "A", Price: 100, Nights: 1},
				},
				Stats: struct {
					ProvidersTotal     int    "json:\"providers_total\""
					ProvidersSucceeded int    "json:\"providers_succeeded\""
					ProvidersFailed    int    "json:\"providers_failed\""
					Cache              string "json:\"cache\""
					DurationMs         int64  "json:\"duration_ms\""
				}{ProvidersTotal: 1, ProvidersSucceeded: 1, ProvidersFailed: 0, Cache: "miss", DurationMs: 50},
			}, nil
		},
	}

	agg := &mockAggregator{}
	rl := &mockRateLimiter{allowFunc: func(ip string) bool { return true }}
	metrics := obs.NewMetrics(prometheus.NewRegistry())

	h := ht.NewHandler(agg, cache, rl, metrics)

	req := httptest.NewRequest("GET", "/search?city=kota&checkin=2025-11-20&nights=2&adults=2", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()

	h.Search(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandler_Search_ValidationFailures(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantCode int
	}{
		{"MissingCity", "?checkin=2025-01-01&nights=2&adults=2", http.StatusBadRequest},
		{"InvalidCheckin", "?city=abc&checkin=2025/01/01&nights=2&adults=2", http.StatusBadRequest},
		{"NightsNotNumber", "?city=abc&checkin=2025-01-01&nights=x&adults=2", http.StatusBadRequest},
		{"NightsZero", "?city=abc&checkin=2025-01-01&nights=0&adults=2", http.StatusBadRequest},
		{"AdultsZero", "?city=abc&checkin=2025-01-01&nights=2&adults=0", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockCache{}
			agg := &mockAggregator{}
			rl := &mockRateLimiter{allowFunc: func(ip string) bool { return true }}
			metrics := obs.NewMetrics(prometheus.NewRegistry())
			h := ht.NewHandler(agg, cache, rl, metrics)

			req := httptest.NewRequest("GET", "/search"+tt.query, nil)
			req.RemoteAddr = "1.2.3.4:1234"
			w := httptest.NewRecorder()

			h.Search(w, req)
			resp := w.Result()

			if resp.StatusCode != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, resp.StatusCode)
			}
		})
	}
}

func TestHandler_Search_RateLimit(t *testing.T) {
	cache := &mockCache{}
	agg := &mockAggregator{}
	rl := &mockRateLimiter{allowFunc: func(ip string) bool { return false }}
	metrics := obs.NewMetrics(prometheus.NewRegistry())

	h := ht.NewHandler(agg, cache, rl, metrics)

	req := httptest.NewRequest("GET", "/search?city=abc&checkin=2025-01-01&nights=2&adults=2", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()

	h.Search(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
}

func TestHandler_Search_AggregatorError(t *testing.T) {
	cache := &mockCache{
		getOrComputeFunc: func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
			return fn(ctx)
		},
	}
	agg := &mockAggregator{
		searchFunc: func(ctx context.Context, req *models.SearchRequest) (search.AggregatedResult, error) {
			return search.AggregatedResult{}, errors.New("aggregator failed")
		},
	}
	rl := &mockRateLimiter{allowFunc: func(ip string) bool { return true }}
	metrics := obs.NewMetrics(prometheus.NewRegistry())

	h := ht.NewHandler(agg, cache, rl, metrics)

	req := httptest.NewRequest("GET", "/search?city=abc&checkin=2025-01-01&nights=2&adults=2", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()

	h.Search(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestHandler_Search_CacheHit(t *testing.T) {
	called := false
	cache := &mockCache{
		getOrComputeFunc: func(ctx context.Context, key string, fn func(ctx context.Context) (search.AggregatedResult, error)) (search.AggregatedResult, error) {
			called = true
			return search.AggregatedResult{
				Hotels: []search.Hotel{{HotelID: "H1", Name: "A", Price: 50, Nights: 1}},
				Stats: struct {
					ProvidersTotal     int    "json:\"providers_total\""
					ProvidersSucceeded int    "json:\"providers_succeeded\""
					ProvidersFailed    int    "json:\"providers_failed\""
					Cache              string "json:\"cache\""
					DurationMs         int64  "json:\"duration_ms\""
				}{ProvidersTotal: 1, ProvidersSucceeded: 1, ProvidersFailed: 0, Cache: "hit"},
			}, nil
		},
	}

	agg := &mockAggregator{}
	rl := &mockRateLimiter{allowFunc: func(ip string) bool { return true }}
	metrics := obs.NewMetrics(prometheus.NewRegistry())

	h := ht.NewHandler(agg, cache, rl, metrics)

	req := httptest.NewRequest("GET", "/search?city=abc&checkin=2025-01-01&nights=1&adults=1", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()

	h.Search(w, req)
	resp := w.Result()

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// basic JSON check
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// check search info
	searchMap, ok := out["search"].(map[string]any)
	if !ok {
		t.Fatal("missing search field")
	}
	if searchMap["city"] != "abc" {
		t.Errorf("expected city abc, got %v", searchMap["city"])
	}

	// check hotels
	hotels, ok := out["hotels"].([]any)
	if !ok || len(hotels) != 1 {
		t.Fatalf("expected 1 hotel, got %+v", hotels)
	}
	hotel := hotels[0].(map[string]any)
	if hotel["hotel_id"] != "H1" {
		t.Errorf("expected hotel_id H1, got %v", hotel["hotel_id"])
	}
	if !called {
		t.Fatal("expected cache GetOrCompute to be called")
	}
}

// You can add more table-driven tests for request ID, malformed RemoteAddr, etc.
