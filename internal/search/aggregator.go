package search

import (
	"context"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/models"
	"github.com/example/mini-hotel-aggregator/internal/obs"
)

type AggregatorService interface {
	Search(ctx context.Context, req *models.SearchRequest) (AggregatedResult, error)
}

// Aggregator queries providers in parallel and merges results.
type Aggregator struct {
	providers []Provider
	timeout   time.Duration
	metrics   *obs.Metrics
}

func NewAggregator(providers []Provider, timeout time.Duration, m *obs.Metrics) *Aggregator {
	return &Aggregator{providers: providers, timeout: timeout, metrics: m}
}

func normalizeHotel(h Hotel) (Hotel, bool) {
	
	h.HotelID = strings.TrimSpace(h.HotelID)
	if h.HotelID == "" || h.Price <= 0 {
		return h, false
	}
	// normalize casing for city
	h.City = strings.ToLower(strings.TrimSpace(h.City))
	return h, true
}

func (a *Aggregator) Search(ctx context.Context, req *models.SearchRequest) (AggregatedResult, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	resCh := make(chan ProviderResult, len(a.providers))
	errCh := make(chan struct{}, len(a.providers)) // only count failures
	var wg sync.WaitGroup
	for _, p := range a.providers {
		wg.Add(1)
		prov := p
		go func(pr Provider) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					log.Printf("provider %s panic recovered: %v", pr.Name(), r)
					a.metrics.IncProviderFailure(pr.Name())
					// non-blocking signal of failure
					select {
					case errCh <- struct{}{}:
					default:
					}
				}
			}()
			start := time.Now()
			hs, err := pr.Search(ctx, req)
			duration := time.Since(start).Seconds()
			a.metrics.ObserveProviderLatency(pr.Name(), duration)

			if err != nil {
				a.metrics.IncProviderFailure(pr.Name())
				// non-blocking send
				select {
				case errCh <- struct{}{}:
				default:
				}
				return
			}
			select {
			case resCh <- ProviderResult{Provider: pr.Name(), Hotels: hs}:
			case <-ctx.Done():
				// context canceled; drop
			}
		}(prov)
	}

	go func() {
		wg.Wait()
		close(resCh)
		close(errCh)
	}()

	all := map[string]Hotel{}
	providersSucceeded := 0
	providersFailed := 0
	// Collect until channels closed or context done
	for resCh != nil || errCh != nil {
		select {
		case pr, ok := <-resCh:
			if !ok {
				resCh = nil
				continue
			}
			providersSucceeded++
			for _, h := range pr.Hotels {
				nh, ok := normalizeHotel(h)
				if !ok {
					continue
				}
				if existing, found := all[nh.HotelID]; !found || nh.Price < existing.Price {
					all[nh.HotelID] = nh
				}
			}
		case _, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			providersFailed++
		case <-ctx.Done():
			// treat remaining as failed
			if resCh != nil || errCh != nil {
				// count remaining providers that didn't respond as failures
				remaining := len(a.providers) - (providersSucceeded + providersFailed)
				if remaining > 0 {
					providersFailed += remaining
				}
			}
			resCh = nil
			errCh = nil
		}
	}

	// build result list
	hotels := make([]Hotel, 0, len(all))
	for _, v := range all {
		hotels = append(hotels, v)
	}
	sort.Slice(hotels, func(i, j int) bool { return hotels[i].Price < hotels[j].Price })

	out := AggregatedResult{}
	out.Stats.ProvidersTotal = len(a.providers)
	out.Stats.ProvidersSucceeded = providersSucceeded
	out.Stats.ProvidersFailed = providersFailed
	out.Stats.Cache = "miss"
	out.Stats.DurationMs = time.Since(start).Milliseconds()
	out.Hotels = hotels
	return out, nil
}
