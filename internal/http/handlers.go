package http

import (
	"net"
	"net/http"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/obs"
	"github.com/example/mini-hotel-aggregator/internal/search"
	"github.com/google/uuid"
)

type Handler struct {
	agg            search.AggregatorService
	cache          search.CacheService
	ratelimiter    search.RateLimiter
	metrics        *obs.Metrics
	computeTimeout time.Duration
}

func NewHandler(agg search.AggregatorService, cache search.CacheService, rl search.RateLimiter, m *obs.Metrics) *Handler {
	return &Handler{agg: agg, cache: cache, ratelimiter: rl, metrics: m, computeTimeout: 3 * time.Second}
}

func (h *Handler) ipFromRequest(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.metrics.IncRequests()

	// chi's middleware.RequestID sets X-Request-Id header
	reqID := r.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = uuid.New().String()
	}

	q := r.URL.Query()
	req, err := models.NewSearchRequest(
		q.Get("city"),
		q.Get("checkin"),
		q.Get("nights"),
		q.Get("adults"),
	)
	if err != nil {
		BadRequest(w, err.Error(), map[string]string{"request_id": reqID})
		return
	}

	if err := req.Validate(); err != nil {
		BadRequest(w, err.Error(), map[string]string{"request_id": reqID})
		return
	}

	// rate limit
	ip := h.ipFromRequest(r)
	if !h.ratelimiter.Allow(ip) {
		h.metrics.IncRateLimitDrops()
		TooManyRequests(w, "rate limit exceeded", map[string]string{"request_id": reqID})
		return
	}

	//passing request to service
	service := search.NewService(h.agg, h.cache, h.metrics, h.computeTimeout)
	res, err := service.Search(ctx, req)
	if err != nil {
		InternalError(w, err.Error(), map[string]string{"request_id": reqID})
		return
	}

	out := map[string]any{
		"search": map[string]any{"city": req.City, "checkin": req.Checkin, "nights": req.Nights, "adults": req.Adults},
		"stats":  res.Stats,
		"hotels": res.Hotels,
	}

	WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
