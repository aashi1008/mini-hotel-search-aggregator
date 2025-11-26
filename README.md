# Mini Hotel Search Aggregator

A Go (1.21+) HTTP service for a take-home assignment. It queries three mock providers in parallel,
merges results, caches responses, rate-limits clients, and exposes health & metrics endpoints.

## Features
- `GET /search?city=...&checkin=YYYY-MM-DD&nights=<int>&adults=<int>`
- Three mock providers (simulated latency/failure)
- Concurrent provider queries with timeouts and context cancellation
- In-memory caching (30s) with request coalescing (collapse concurrent requests)
- Rate limit: 10 requests/min per IP (in-memory token bucket)
- Observability: `/healthz`, `/metrics` with basic counters (plain text)
- Tests for merging, cache collapse, and rate limiter

## Run (development)
Requires Go 1.21+ and network to fetch chi router.

```bash
git clone <repo>
cd mini-hotel-aggregator
go mod download
go run ./cmd/server
```

Server runs on `:8080` by default. See `Makefile` for convenience commands.

## Example curl
```bash
curl 'http://localhost:8080/search?city=marrakesh&checkin=2025-11-20&nights=2&adults=2'
```

## Project Structure
- `/cmd/server/main.go` - entrypoint and router
- `/internal/http` - HTTP handlers
- `/internal/search` - aggregator, cache, ratelimiter, types
- `/internal/providers` - three mock providers
- `/internal/obs` - simple metrics collector

## Notes
- Prices assumed same currency.
- This is a self-contained in-memory prototype for the take-home test.
