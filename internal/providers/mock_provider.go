package providers

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/search"
)

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

func (m *MockProvider) Search(ctx context.Context, city, checkin string, nights, adults int) ([]search.Hotel, error) {
	// variable latency and context cancelable
	select {
	case <-time.After(sampleLatencyFromRng(m.rng, m.avgLatency)):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if shouldFailFromRng(m.rng, m.failRate) {
		return nil, errors.New("provider error (simulated)")
	}

	hotels := []search.Hotel{
		{HotelID: "H123", Name: "Hotel Atlas", City: city, Currency: "EUR", Price: 129.90 + float64(m.rng.Intn(30)), Nights: nights},
		{HotelID: "H234", Name: "Riad Sunset", City: city, Currency: "EUR", Price: 99.50 + float64(m.rng.Intn(100)), Nights: nights},
		{HotelID: "H345", Name: "Kasbah Pearl", City: city, Currency: "EUR", Price: 132.00 + float64(m.rng.Intn(40)), Nights: nights},
	}
	b, _ := json.Marshal(hotels)
	var out []search.Hotel
	_ = json.Unmarshal(b, &out)
	return out, nil
}
