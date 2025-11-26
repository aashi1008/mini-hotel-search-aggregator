package providers

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/search"
)

type Provider interface {
    Search(ctx context.Context, req *models.SearchRequest) ([]search.Hotel, error)
    Name() string
}

type MockProvider struct {
	name       string
	avgLatency float64
	failRate   float64
	rng        *rand.Rand
}

func NewMockProvider(name string, avgLatency, failRate float64, seedOffset int64) *MockProvider {
	seed := time.Now().UnixNano() + seedOffset
	return &MockProvider{name: name, avgLatency: avgLatency, failRate: failRate, rng: rand.New(rand.NewSource(seed))}
}

func (m *MockProvider) Name() string { return m.name }

func (m *MockProvider) Search(ctx context.Context, req *models.SearchRequest) ([]search.Hotel, error) {
	// variable latency and context cancelable
	select {
	case <-time.After(SampleLatencyFromRng(m.rng, m.avgLatency)):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if ShouldFailFromRng(m.rng, m.failRate) {
		return nil, errors.New("provider error (simulated)")
	}

	hotels := []search.Hotel{
		{HotelID: "H123", Name: "Hotel Atlas", City: req.City, Currency: "EUR", Price: 129.90 + float64(m.rng.Intn(30)), Nights: req.Nights},
		{HotelID: "H234", Name: "Riad Sunset", City: req.City, Currency: "EUR", Price: 99.50 + float64(m.rng.Intn(100)), Nights: req.Nights},
		{HotelID: "H345", Name: "Kasbah Pearl", City: req.City, Currency: "EUR", Price: 132.00 + float64(m.rng.Intn(40)), Nights: req.Nights},
	}

	return hotels, nil
}

func SampleLatencyFromRng(rng *rand.Rand, avg float64) time.Duration {
	ms := float64(50) + rng.ExpFloat64()*avg*200.0
	return time.Duration(ms) * time.Millisecond
}

func ShouldFailFromRng(rng *rand.Rand, rate float64) bool {
	return rng.Float64() < rate
}
