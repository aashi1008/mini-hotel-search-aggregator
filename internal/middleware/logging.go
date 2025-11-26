package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// LoggingMiddleware logs basic request info (request-id, method, path, status, duration).
func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get("X-Request-Id")
			if rid == "" {
				if v := r.Context().Value(middleware.RequestIDKey); v != nil {
					if s, ok := v.(string); ok {
						rid = s
					}
				}
			}
			start := time.Now()
			rec := &statusRecorder{
				ResponseWriter: w,
				Status:         http.StatusOK, // default until changed
			}
			next.ServeHTTP(rec, r)
			duration := time.Since(start)
			
			logger.Info("request completed",
				"request_id", rid,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.Status,
				"duration_ms", strconv.FormatInt(duration.Milliseconds(), 10),
			)
		}
		return http.HandlerFunc(fn)
	}
}
