package obs

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestsTotal       prometheus.Counter
	CacheHitsTotal      prometheus.Counter
	RateLimitDropsTotal prometheus.Counter

	ProviderErrors      *prometheus.CounterVec
	ProviderLatency     *prometheus.HistogramVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsTotal   *prometheus.CounterVec
	Registry            *prometheus.Registry
}

// Create Prometheus collectors and register them
func NewMetrics(p *prometheus.Registry) *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hotel_requests_total",
			Help: "Total number of incoming search requests",
		}),
		CacheHitsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hotel_cache_hits_total",
			Help: "Number of cache hits for search results",
		}),
		ProviderErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "provider_errors_total",
			Help: "Errors returned by each provider",
		}, []string{"provider"},
		),
		RateLimitDropsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "hotel_ratelimit_drops_total",
			Help: "Requests dropped due to rate limiting",
		}),
		ProviderLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "provider_latency_ms",
				Help:    "Latency between aggregator and provider",
				Buckets: prometheus.LinearBuckets(5, 20, 15),
			},
			[]string{"provider"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latencies",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		Registry: p,
	}

	// Register metrics with Prometheus
	p.MustRegister(
		m.RequestsTotal,
		m.CacheHitsTotal,
		m.ProviderErrors,
		m.RateLimitDropsTotal,
		m.ProviderLatency,
		m.HTTPRequestDuration,
		m.HTTPRequestsTotal,
	)

	return m
}

func (m *Metrics) IncRequests()  { m.RequestsTotal.Inc() }
func (m *Metrics) IncCacheHits() { m.CacheHitsTotal.Inc() }

func (m *Metrics) IncRateLimitDrops() { m.RateLimitDropsTotal.Inc() }

func (m *Metrics) ObserveProviderLatency(provider string, seconds float64) {
	m.ProviderLatency.WithLabelValues(provider).Observe(seconds)
}

func (m *Metrics) IncProviderFailure(provider string) {
	m.ProviderErrors.WithLabelValues(provider).Inc()
}

func (m *Metrics) ObserveHTTPRequestDuration(method string, path string, status string, seconds float64) {
	m.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(seconds)
}

func (m *Metrics) IncHTTPRequestsTotal(method string, path string, status string) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}
