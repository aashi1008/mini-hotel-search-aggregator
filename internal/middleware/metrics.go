package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/obs"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(m *obs.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, Status: 200}

			next.ServeHTTP(rec, r)

			status := strconv.Itoa(rec.Status)
			method := r.Method
			path := r.URL.Path

			m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
			m.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(time.Since(start).Seconds())
		}

		return http.HandlerFunc(fn)
	}
}
