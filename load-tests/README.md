# Load tests (k6)

Realistic user-behavior scripts. Each VU signs up + logs in on its first
iteration (per-VU unique email), then drives a flow with think-time sleeps.
Metrics stream to Mimir via Prometheus remote-write and show on the
`k6-load-test` Grafana dashboard.

## Scripts

| Script | Profile | What a VU does |
|---|---|---|
| `smoke.js` | 1 VU × 30s | Touches every endpoint once. Sanity check. |
| `browse.js` | ramping VUs | Lists products, scrolls a few pages, opens 1–2 detail pages. |
| `cart.js` | ramping VUs | Browse + adds 1–3 items, may view cart, may remove one, abandons. |
| `checkout.js` | ramping VUs | Browse + 1–2 items + `/checkout` + `/orders/{id}`. |

No SLO thresholds yet — k6 always exits 0. Judge results in Grafana.

## Running

Prereqs:
1. App stack: `make run-catalog`, `make run-cart`, `make run-order`, `make run-gateway` (one per terminal).
2. Observability: `docker compose -f observability/docker-compose.yml up -d`.

Then:

```bash
make load-smoke                                  # quick sanity
make load-browse                                 # default: 100 VUs × 3m
make load-checkout K6_VUS=1000 K6_DURATION=5m    # 1k VUs
make load-checkout K6_VUS=10000 K6_DURATION=10m  # 10k VUs
```

## What to watch in Grafana while a test runs

- `k6-load-test` dashboard: `k6_vus` should ramp to the target, `k6_http_reqs_total` rate climbs, `k6_http_req_failed_rate` stays low (expect ~20% on `checkout.js` because the fake payment provider declines 20% by design).
- `service-overview` dashboard: app-side RPS, p95 latency, error rate per service. When the system stays healthy the app's p95 should track k6's p99 closely; when it doesn't, you've found a bottleneck.

## Scaling above 1k VUs

10k VUs from a laptop is heavy. Known issues and workarounds:

- **Ephemeral port exhaustion** (Windows/macOS) — defaults around 5k. Symptoms: errors mentioning `bind: An attempt was made to access a socket in a way forbidden by its access permissions`. Fix: lower the TIME_WAIT timeout (Windows: `netsh int ipv4 set dynamic tcp start=10000 num=55000`) or run from Linux/WSL where the ephemeral range is wider.
- **Open file descriptors** (Linux) — `ulimit -n 65536` before running k6.
- **DB connection pool** — at 10k VUs each gateway request opens a transaction. If you see `too many connections` in Postgres logs, raise `DATABASE_MAX_CONNS` (see `internal/database/pool.go`) and tune `max_connections` in `postgres.conf`.
- **Mimir ingestion rate** — the `mimir-config.yaml` has `ingestion_rate: 100000`. If k6 starts dropping metric pushes (look in the collector logs), bump it.

For distributed load above what one machine can produce, see k6's docs on `k6-operator` (Kubernetes) or k6 Cloud. Out of scope here.

## How user accounts work

Every VU signs up exactly once (on its first iteration) with an email like
`loadtest_<runId>_<__VU>@test.local`. `runId` is a per-run nonce so two runs
never collide on the unique-email constraint. There is no cleanup — these
users accumulate in the `users` table. Manual purge:

```sql
DELETE FROM users WHERE email LIKE 'loadtest_%@test.local';
```
