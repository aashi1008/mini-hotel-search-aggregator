package search

import (
	"context"
	"testing"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/obs"
	"github.com/prometheus/client_golang/prometheus"
)

// basic test using deterministic provider
type staticProvider struct {
	name   string
	hotels []Hotel
}

func (s *staticProvider) Search(ctx context.Context, req *models.SearchRequest) ([]Hotel, error) {
	return s.hotels, nil
}
func (s *staticProvider) Name() string { return s.name }

func TestAggregatorMerge(t *testing.T) {
	p1 := &staticProvider{"p1", []Hotel{{HotelID: "H1", Name: "A", Price: 100, Nights: 1}}}
	p2 := &staticProvider{"p2", []Hotel{{HotelID: "H1", Name: "A", Price: 90, Nights: 1}, {HotelID: "H2", Name: "B", Price: 150, Nights: 1}}}
	agg := NewAggregator([]Provider{p1, p2}, 1*time.Second, obs.NewMetrics(prometheus.NewRegistry()))
	ctx := context.Background()
	req := &models.SearchRequest{
		City:    "city",
		Checkin: "2025-11-20",
		Nights:  1,
		Adults:  2,
	}
	res, err := agg.Search(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Hotels) != 2 {
		t.Fatalf("expected 2 hotels got %d", len(res.Hotels))
	}
	if res.Hotels[0].HotelID != "H1" || res.Hotels[0].Price != 90 {
		t.Fatalf("expected H1 with price 90 got %+v", res.Hotels[0])
	}
}

func TestAggregator_NormalizesAndDeduplicates(t *testing.T) {
	providers := []Provider{
		&staticProvider{"p1", []Hotel{
			{HotelID: "H1", Name: "Hotel X", Price: 80, City: "NYc"},
			{HotelID: "H2", Name: "Hotel Y", Price: -10, City: "NYC"}, // negative price, should drop
			{HotelID: "", Name: "Hotel Z", Price: 120, City: "NYC"},   // missing ID, should drop
		}},
		&staticProvider{"p2", []Hotel{
			{HotelID: "H1", Name: "Hotel X", Price: 75, City: "nyc"}, // duplicate ID, use lowest price
			{HotelID: "H3", Name: "Hotel W", Price: 150, City: "NYC"},
		}},
	}
	req := &models.SearchRequest{
		City:    "NYc",
		Checkin: "2025-11-20",
		Nights:  2,
		Adults:  2,
	}
	agg := NewAggregator(providers, 1*time.Second, obs.NewMetrics(prometheus.NewRegistry()))
	res, err := agg.Search(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Hotels) != 2 {
		t.Fatalf("expected 2 hotels, got %d", len(res.Hotels))
	}
	if res.Hotels[0].HotelID != "H1" || res.Hotels[0].Price != 75 {
		t.Errorf("expected H1 with price 75, got %+v", res.Hotels[0])
	}
	// assert dropped invalid/negative/missing ID hotels
	for _, h := range res.Hotels {
		if h.Price <= 0 || h.HotelID == "" {
			t.Errorf("should not have invalid hotel %+v", h)
		}
		if h.City != "nyc" {
			t.Errorf("city should be normalized to lowercase, got %s", h.City)
		}
	}
}
