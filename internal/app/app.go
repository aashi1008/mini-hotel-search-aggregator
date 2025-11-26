package app

import (
	"net/http"
	"time"

	handlers "github.com/example/mini-hotel-aggregator/internal/http"
	"github.com/example/mini-hotel-aggregator/internal/obs"
	"github.com/example/mini-hotel-aggregator/internal/providers"
	"github.com/example/mini-hotel-aggregator/internal/routes"
	"github.com/example/mini-hotel-aggregator/internal/search"
	"github.com/prometheus/client_golang/prometheus"
)

type App struct {
	Router      http.Handler
	Aggregator  *search.Aggregator
	Cache       *search.Cache
	RateLimiter *search.IPRateLimiter
	Metrics     *obs.Metrics
}

func SetAppConfig() *App {
	providersList := []search.Provider{
		providers.NewMockProvider("mock1", 0.2, 0.10, 0),
		providers.NewMockProvider("mock2", 0.25, 0.12, 1),
		providers.NewMockProvider("mock3", 0.15, 0.05, 2),
	}

	customRegistry := prometheus.NewRegistry()
	metrics := obs.NewMetrics(customRegistry)
	agg := search.NewAggregator(providersList, 2*time.Second, metrics)
	cache := search.NewCache(30 * time.Second, metrics)
	rl := search.NewIPRateLimiter(10, time.Minute)
	h := handlers.NewHandler(agg, cache, rl, metrics)

	router := routes.GetRoutes(h, metrics)

	return &App{
		Router:      router,
		Aggregator:  agg,
		Cache:       cache,
		RateLimiter: rl,
		Metrics:     metrics,
	}
}
