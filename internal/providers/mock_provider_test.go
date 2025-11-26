package providers_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/providers"
)

// helper to create a deterministic MockProvider
func newTestProvider() *providers.MockProvider {
	return providers.NewMockProvider("mock1", 0.1, 0.0, 0)
}

func TestMockProvider_Search_Positive(t *testing.T) {
	p := newTestProvider()

	req := &models.SearchRequest{
		City:    "Paris",
		Checkin: "2025-12-01",
		Nights:  2,
		Adults:  2,
	}

	ctx := context.Background()
	hotels, err := p.Search(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(hotels) != 3 {
		t.Fatalf("expected 3 hotels, got %d", len(hotels))
	}

	for _, h := range hotels {
		if h.City != "Paris" {
			t.Errorf("expected city Paris, got %s", h.City)
		}
		if h.Nights != 2 {
			t.Errorf("expected nights 2, got %d", h.Nights)
		}
		if h.Price <= 0 {
			t.Errorf("expected price > 0, got %f", h.Price)
		}
	}
}

func TestMockProvider_Search_Failure(t *testing.T) {
	p := providers.NewMockProvider("mock-fail", 0.0, 1.0, 0) // failRate 100%

	req := &models.SearchRequest{City: "Paris", Checkin: "2025-12-01", Nights: 2, Adults: 2}
	ctx := context.Background()
	_, err := p.Search(ctx, req)
	if err == nil {
		t.Fatal("expected an error due to failRate 100%")
	}
	if err.Error() != "provider error (simulated)" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockProvider_Search_ContextCancelled(t *testing.T) {
	p := newTestProvider()

	req := &models.SearchRequest{City: "Paris", Checkin: "2025-12-01", Nights: 2, Adults: 2}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := p.Search(ctx, req)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
}

func TestSampleLatencyFromRng(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	d := providers.SampleLatencyFromRng(rng, 0.1)
	if d <= 0 {
		t.Errorf("expected positive latency, got %v", d)
	}
}

func TestShouldFailFromRng(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	count := 0
	for i := 0; i < 1000; i++ {
		if providers.ShouldFailFromRng(rng, 0.5) {
			count++
		}
	}
	if count == 0 || count == 1000 {
		t.Errorf("expected some failures with 50%% rate, got %d/1000", count)
	}
}
