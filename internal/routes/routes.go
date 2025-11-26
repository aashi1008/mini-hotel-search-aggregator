package routes

import (
	"time"

	handlers "github.com/example/mini-hotel-aggregator/internal/http"
	mid "github.com/example/mini-hotel-aggregator/internal/middleware"
	"github.com/example/mini-hotel-aggregator/internal/obs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func GetRoutes(h *handlers.Handler, metrics *obs.Metrics) *chi.Mux{
	r := chi.NewRouter()
	// Useful built-in middlewares
	r.Use(mid.MetricsMiddleware(metrics))
	r.Use(middleware.RealIP)    // proper client IP extraction
	r.Use(middleware.RequestID) // sets request ID header
	r.Use(middleware.Recoverer) // built-in recoverer to avoid panics taking server down

	// our custom middlewares: logging & timeout
	r.Use(mid.LoggingMiddleware)
	r.Use(mid.TimeoutMiddleware(10 * time.Second))

	// endpoints
	r.Get("/search", h.Search)
	r.Get("/healthz", h.Healthz)
	r.Get("/metrics", metrics.Handler().ServeHTTP)
	//r.Handle("/metrics", promhttp.HandlerFor(metrics.CustomRegistry, promhttp.HandlerOpts{}))

	return r
}
