package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// LoggingMiddleware logs basic request info (request-id, method, path, duration).
func LoggingMiddleware(next http.Handler) http.Handler {
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
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", rid, r.Method, r.URL.Path, time.Since(start))
	}
	return http.HandlerFunc(fn)
}
