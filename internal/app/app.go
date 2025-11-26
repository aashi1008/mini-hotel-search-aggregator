package app

import (
	"log/slog"
	"net/http"
	"os"
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
	Aggregator  search.AggregatorService
	Cache       search.CacheService
	RateLimiter search.RateLimiter
	Metrics     *obs.Metrics
}

func SetAppConfig() *App {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	providersList := []search.Provider{
		providers.NewMockProvider("mock1", 0.2, 0.10, 0),
		providers.NewMockProvider("mock2", 0.25, 0.12, 1),
		providers.NewMockProvider("mock3", 0.15, 0.05, 2),
	}

	customRegistry := prometheus.NewRegistry()
	metrics := obs.NewMetrics(customRegistry)
	agg := search.NewAggregator(providersList, 2*time.Second, metrics)
	cache := search.NewCache(30*time.Second, metrics)
	rl := search.NewIPRateLimiter(10, time.Minute)
	h := handlers.NewHandler(agg, cache, rl, metrics)

	router := routes.GetRoutes(h, metrics, logger)

	return &App{
		Router:      router,
		Aggregator:  agg,
		Cache:       cache,
		RateLimiter: rl,
		Metrics:     metrics,
	}
}
