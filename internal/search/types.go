package search

import "context"

type Hotel struct {
	HotelID  string  `json:"hotel_id"`
	Name     string  `json:"name"`
	City     string  `json:"city"`
	Currency string  `json:"currency"`
	Price    float64 `json:"price"`
	Nights   int     `json:"nights"`
}

type ProviderResult struct {
	Provider string
	Hotels   []Hotel
}

type AggregatedResult struct {
	Stats struct {
		ProvidersTotal     int    `json:"providers_total"`
		ProvidersSucceeded int    `json:"providers_succeeded"`
		ProvidersFailed    int    `json:"providers_failed"`
		Cache              string `json:"cache"`
		DurationMs         int64  `json:"duration_ms"`
	} `json:"stats"`
	Hotels []Hotel `json:"hotels"`
}

type Provider interface {
	Search(ctx context.Context, city, checkin string, nights, adults int) ([]Hotel, error)
	Name() string
}