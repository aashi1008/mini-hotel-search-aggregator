package search

import (
	"context"
	"fmt"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/obs"
)

type Service struct {
	agg            *Aggregator
	cache          *Cache
	metrics        *obs.Metrics
	computeTimeout time.Duration
}

func NewService(ag *Aggregator, ch *Cache, m *obs.Metrics, t time.Duration) *Service {
	return &Service{
		agg:            ag,
		cache:          ch,
		metrics:        m,
		computeTimeout: t,
	}
}

func (s *Service) Search(ctx context.Context, req *models.SearchRequest) (AggregatedResult, error) {
	cacheKey := fmt.Sprintf("%s|%s|%d|%d", req.City, req.Checkin, req.Nights, req.Adults)

	// compute with per-request timeout
	cctx, cancel := context.WithTimeout(ctx, s.computeTimeout)
	defer cancel()

	res, cacheHit := s.cache.GetOrCompute(cctx, cacheKey, func(ctx context.Context) (AggregatedResult, error) {
		return s.agg.Search(ctx, req)
	})
	if cacheHit {
		s.metrics.IncCacheHits()
	}

	return res, nil
}
