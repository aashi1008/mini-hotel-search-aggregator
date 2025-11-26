package search

import (
	"context"
	"fmt"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/obs"
)

type ServiceManagement interface{
	Search(ctx context.Context, req *models.SearchRequest) (AggregatedResult, error)
}

type service struct {
	agg            AggregatorService
	cache          CacheService
	metrics        *obs.Metrics
	computeTimeout time.Duration
}

func NewService(ag AggregatorService, ch CacheService, m *obs.Metrics, t time.Duration) *service {
	return &service{
		agg:            ag,
		cache:          ch,
		metrics:        m,
		computeTimeout: t,
	}
}

func (s *service) Search(ctx context.Context, req *models.SearchRequest) (AggregatedResult, error) {
	cacheKey := fmt.Sprintf("%s|%s|%d|%d", req.City, req.Checkin, req.Nights, req.Adults)

	// compute with per-request timeout
	cctx, cancel := context.WithTimeout(ctx, s.computeTimeout)
	defer cancel()

	res, err := s.cache.GetOrCompute(cctx, cacheKey, func(ctx context.Context) (AggregatedResult, error) {
		return s.agg.Search(ctx, req)
	})
	if err != nil {
		return AggregatedResult{}, err
	}

	return res, nil
}
