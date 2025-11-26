
# 🏨 Mini Hotel Aggregator

A lightweight, concurrency-optimized hotel search aggregator written in Go.  
Queries multiple providers in parallel, merges results, caches responses with singleflight, rate-limits abusive clients, and exposes Prometheus metrics.

## 🚀 Features

- Parallel provider requests (fan-out / fan-in via goroutines)
- Deterministic and randomized mock providers
- Request-level caching with singleflight (prevents cache stampede)
- Token bucket IP-based rate limiting
- Structured request validation via DTOs
- Prometheus metrics: request count, cache hits, durations, rate limit drops
- Health check endpoint
- Structured logging with request ID, duration
- Graceful shutdown (context cancellation) & Recover panic through chi framework
- Comprehensive mocks and unit tests for all core modules

## 📦 Project Structure

---
```
internal/
  app/          # Dependency wiring, SetAppConfig
  http/         # Handlers and request DTOs
  routes/       # Router initialization
  search/       # Aggregator, cache, rate limiter, types
  providers/    # Mock providers, easy extensibility
  models/       # Shared types (SearchRequest, Hotel, etc)
  obs/          # Prometheus metrics instrumentation
  validator/    # Validating mandatory request fields
cmd/
  server/       # Entry point (main.go)
Makefile, README.md, go.mod, go.sum
```
---
## ▶️ Getting Started

### 1. Clone the repo and install dependencies
---
---
```sh
git clone https://github.com/example/mini-hotel-aggregator.git
cd mini-hotel-aggregator
go mod tidy
```
---
### 2. Run the server
---
---
```sh
go run ./cmd/server
```
---
Server runs on: [http://localhost:8080](http://localhost:8080)

### 3. Run with Docker
---

---
```go
docker build -t mini-hotel-aggregator .

docker run -d \
  -p 8080:8080 \
  --name mini-hotel mini-hotel-aggregator
```
---

## 🔍 API Reference

### 1. Search Hotels

**Endpoint:** `GET /search?city=...&checkin=YYYY-MM-DD&nights=N&adults=N`

**Example:**
---
---
```sh
curl "http://localhost:8080/search?city=marrakesh&checkin=2025-11-20&nights=2&adults=2"
```
---
**Sample Response:**
---
---
```json
{
  "search":  {"city":"marrakesh","checkin":"2025-11-20","nights":2,"adults":2},
  "stats": {"providers_total":3,"providers_succeeded":2,"providers_failed":1,"cache":"miss","duration_ms":412},
  "hotels": [
    {"hotel_id": "H123", "name": "Hotel Atlas", "currency": "EUR", "price": 129.9}
  ]
}
```
---
### 2. Health Check

---
```sh
curl http://localhost:8080/healthz
```
---
### 3. Prometheus Metrics

---
```sh
curl http://localhost:8080/metrics
```
---
Metrics scraped by Prometheus and visualized in Grafana.

## 🧠 Design & Architecture

### Request DTO Pattern

- Request mapped to DTO: validation, sanitization, domain separation.
- Promotes testable, maintainable code.

### Aggregator (fan-out, fan-in)

- All providers queried in parallel.
- Context-based timeouts.
- Provider failures logged; successful results merged/sorted/deduped by hotel ID.

### Cache (Singleflight + TTL)

- Prevents stampede: concurrent identical requests share a single computation.
- TTL controls cache freshness.
- Returns stale data only if computation in progress or errors.

### Rate Limiting

- Per-IP token bucket (default 10/min).
- Excess requests return HTTP 429; metrics incremented.

### Metrics & Observability

- Prometheus metrics:
  - `http_requests_total`
  - `http_rate_limit_drops_total`
  - `cache_hits_total`
  - `http_request_duration_seconds` (Histogram)
- `/metrics` endpoint for scraping.
- `/healthz` for basic status.
- Structured logs via middleware with request ID and latency.

### Graceful Shutdown

- `SIGINT`/`SIGTERM` signals trigger cancellation of all in-flight operations.
- Clean shutdown with timeout.

### Testing

- Table-driven, deterministic unit tests (aggregator, cache, rate limiter).
- Provider tests use seeded RNG for reproducibility.
- Cache collapse and concurrency safety tested.
- Run `go test -race ./...` before deployment for race condition detection.

## 🔧 Extending the Project

- **Add a real provider:** Implement the `Provider` interface, register in `internal/app/app.go`.
- **New filters or features:** Add logic inside `internal/search/`.
- **Additional metrics:** Register collectors in `internal/obs/metrics.go`.
- **OpenAPI Spec:** (Optional) Document API with Swagger or similar.

## 🗃️ Roadmap & TODO

- Real provider integration.
- Dockerfile & deployment manifest.
- Advanced filtering (price, amenities).
- API versioning.
- More resilient error handling and alerting.
- CI/CD pipeline for automated lint/test/build.

