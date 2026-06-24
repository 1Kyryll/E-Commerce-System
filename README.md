# E-Commerce System

[![CI](https://github.com/1Kyryll/E-Commerce-System/actions/workflows/ci.yml/badge.svg)](https://github.com/1Kyryll/E-Commerce-System/actions/workflows/ci.yml)
[![Contributions welcome](https://img.shields.io/badge/Contributions-welcome-brightgreen.svg?style=flat)](https://github.com/dwyl/esta/issues)

A learning-focused e-commerce system built to practice production concerns end to
end: system design under load, observability, load testing, CI/CD, and cloud
deployment. The purchase flow is designed so that no item is oversold, no payment
is lost or made twice, and no order is silently dropped — even with thousands of
users contending for the same scarce stock.

`docs/` is the source of truth for every design decision. Start there:
[system-design](docs/system-design.md) ·
[code-architecture](docs/code-architecture.md) ·
[data-layer](docs/data-layer.md) ·
[ADRs](docs/adr/) ·
[observability](docs/observability.md) ·
[load-testing](docs/load-testing.md) ·
[pipeline](docs/pipeline.md)

## Tech stack

- **Backend** — Go: gRPC internally, a REST gateway for the browser, `pgx` + `sqlc` for Postgres
- **Frontend** — Next.js (App Router) + TypeScript + Tailwind v4
- **Data** — Postgres (source of truth)
- **Observability** — OpenTelemetry → Grafana LGTM (Loki, Grafana, Tempo, Mimir/Prometheus)
- **Load testing** — Grafana k6
- **CI/CD** — GitHub Actions
- **Deployment** — Docker Compose on Hetzner Cloud

## Architecture

The browser talks REST to a single **gateway**, which fans out over gRPC to the
internal services. Catalog reads go through a Redis cache-aside; everything else
hits Postgres directly. Order placement uses an atomic inventory decrement,
time-bound reservations for the payment window, and a transactional outbox for
async fan-out. See [system-design.md](docs/system-design.md) for the full flow.

| Service | Type | Purpose |
|---|---|---|
| `gateway` | REST (HTTP) | Browser-facing API; translates REST → gRPC, owns auth |
| `catalog` | gRPC | Product catalog (cache-aside over Redis + Postgres) |
| `cart` | gRPC | Persistent, user-scoped cart |
| `order` | gRPC | Order placement: reservations, payment, outbox |
| `cleanup-worker` | worker | Releases expired reservations, restoring inventory |

## Prerequisites

- Docker + Docker Compose
- Go 1.25 and Node 22 (for local dev outside containers)
- The shared Docker network: `docker network create ecom-net`
- Generated backend code (proto + sqlc) before any Go build:
  ```bash
  make generate-backend   # or `make generate` to also regenerate frontend types
  ```

## Quick start (full stack in Docker)

```bash
docker network create ecom-net          # once
make generate-backend                   # produces backend/gen (gitignored)
docker compose up --build               # postgres, redis, migrate, 5 services, frontend
```

A one-shot `migrate` service applies `backend/migrations` before the services
start. Once up:

| What | URL |
|---|---|
| Storefront (frontend) | http://localhost:3000 |
| Gateway API | http://localhost:8080 |
| Postgres | localhost:5433 |
| Redis | localhost:6379 |

Tear down with `docker compose down` (add `-v` to drop the Postgres volume).

## Local development (without app containers)

Run the datastores in containers and the services on the host:

```bash
docker compose up -d postgres redis
cp backend/.env.example backend/.env     # fill in JWT_SECRET etc.
make migrate-up                          # apply migrations
make run-gateway                         # and run-catalog / run-cart / run-order
make frontend-dev                        # next dev server on :3000
```

Service ports and env vars are documented in [backend/.env.example](backend/.env.example).

## Observability

The LGTM stack runs from a dedicated compose file:

```bash
docker compose -f observability/docker-compose.yml up -d
```

Point the services at the collector with `OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317`
in `backend/.env`. Grafana: http://localhost:3001 (anonymous, admin).

### k6 metrics → Grafana

```bash
K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9009/api/v1/push \
K6_PROMETHEUS_RW_PUSH_INTERVAL=5s \
k6 run --out experimental-prometheus-rw load-tests/checkout.js
```

Metrics land in Mimir under `k6_*` series and show on the `k6-load-test` dashboard.

## Project layout

```
proto/           gRPC contracts (buf)
backend/         Go module — cmd/<service>, internal/, services/, gateway/, migrations/
frontend/        Next.js App Router storefront (feature-sliced)
observability/   Grafana LGTM stack (compose + provisioning)
load-tests/      k6 scripts (smoke, browse, cart, checkout)
docs/            Authoritative design docs + ADRs
```

## Tests & CI

```bash
cd backend && go test ./...                          # unit
cd backend && go test -tags=integration ./...        # integration (needs Postgres)
cd frontend && npm run typecheck && npm run lint && npm run build
```

CI runs the backend test job (proto + sqlc codegen, unit + integration tests) and a
frontend job (generated-types check, typecheck, lint, build) on every push.

## TODO 

- [ ] Cart atomic change 
- [ ] List User Products 
- [ ] Redis cache
- [ ] Polish Grafana dashboards
- [ ] Robust load-testing with 10K+ VUs
- [ ] Deploy the application 
- [ ] Load Balancer 
- [ ] Down-stream services(Email, Warehouse, etc.)