# pgx default pool of 4 connections starves under load

## Symptom

While load-testing checkout against a freshly-started stack, `/cart/*` and `/checkout` endpoints failed at ~93% rate even though the dashboards showed Postgres and the application services healthy. k6 reported a mix of `status=0` (request timeouts past the 10s ceiling) and `status=401 unauthorized`. The `service-overview` dashboard showed the cart service request rate climbing but its p95 latency exploding from ~5 ms to multiple seconds the moment VU count crossed ~25. The catalog service stayed at 100% success the whole time, which made the asymmetry stand out.

Cart endpoints failed; product listings stayed perfect. Same gateway, same backend module, same VUs — the only thing that varied was the work each downstream service did per request.

## First hypothesis

Initial assumption was that this was a connection-cookie or session-propagation problem on the load-test side: VUs failing to keep sessions across iterations, so cart calls landed unauthenticated and returned 401. That theory matched the 401 fraction but didn't explain the `status=0` timeouts at all, and only partially explained the asymmetry — catalog endpoints are unauthenticated, so even a VU with a broken session would still get a clean 200 from `/products`. That coincidence pulled attention toward auth and away from the database layer for longer than it should have.

## Diagnosis

The right way to read the asymmetry was *what does each endpoint actually do at the data layer*, not "is one of them cached." Catalog reads in the current state are still served from Postgres (the cache-aside path described in ADR 6 isn't implemented yet), so the comparison wasn't "DB vs Redis" — it was:

| | Catalog (`GET /products`) | Cart (`POST /cart/items`) |
|---|---|---|
| Auth | not required | session cookie required |
| Query shape | single `SELECT` | multi-statement transaction |
| Connection occupancy | brief — return to pool immediately | held for the duration of `BEGIN ... COMMIT` |

Cart's `AddItem` opens a transaction with three statements (`UpsertCart`, `UpsertCartItem`, `ListCartItems`). Each call holds a single pgx connection for the entire transaction round-trip, not just per statement. Catalog's `GET /products` does one short SELECT and releases the connection immediately. Two endpoints that both touch the same DB can have wildly different pool-occupancy profiles.

Three observations narrowed it down:

1. **A single-VU debug script proved auth and cookies were fine.** Running the same flow at concurrency 1 returned `GET /cart → 200` with a non-empty item list. So at low concurrency every layer was healthy; the bug was load-dependent.
2. **`pg_stat_activity` during a concurrent run** showed the cart service holding 4 connections in `active` state, with the rest of its pool at 0 idle — meaning pgx wasn't opening more connections, it was capping itself there.
3. **Reading `pool.go`** showed the project never set `MaxConns` on `pgxpool.Config`. Cross-referencing pgx confirmed its default: `max(runtime.NumCPU(), 4)`. Every service was running with **a pool of 4 connections.** At 100 concurrent VUs the cart service had ~96 transactions queued waiting for a connection; queue wait time stacked past k6's 10-second client timeout, which surfaced as `status=0` errors.

The 401s were a downstream cascade: when signup-on-iter-0 timed out (the gateway's own pool was equally starved when 100 VUs hit `/auth/signup` simultaneously), the VU never got a session cookie, and every subsequent iteration of that VU returned 401 on every authenticated endpoint. Two independent bugs were intertwined; the cookie-jar issue (documented in [k6-cookie-jar.md](./k6-cookie-jar.md)) made the cascade worse but wasn't the trigger. The trigger was the pool.

## Root cause

`pgxpool` defaults `MaxConns` to `max(runtime.NumCPU(), 4)` when the caller doesn't set it explicitly. For an application that opens multi-statement transactions, this collapses under any concurrency above the connection count — every in-flight transaction occupies a connection for its full lifetime, not just per query.

## Fix

Plumb a `DATABASE_MAX_CONNS` environment variable through `internal/database/pool.go` and bump the default to **25 per service**. With 4 services in the system that puts the steady-state ceiling at ~100 connections — which is also Postgres's default `max_connections` — so the server-side cap was raised to **200** in `docker-compose.yml` to leave headroom for `psql` sessions, monitoring, and the inevitable connection leaks from service restarts during development.

```go
// backend/internal/database/pool.go
const DefaultMaxConns int32 = 25

max := cfg.MaxConns
if max <= 0 {
    if v := os.Getenv("DATABASE_MAX_CONNS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            max = int32(n)
        }
    }
}
if max <= 0 {
    max = DefaultMaxConns
}
pcfg.MaxConns = max
```

```yaml
# docker-compose.yml — give Postgres room for 4 × 25 pools + headroom.
postgres:
  command: ["postgres", "-c", "max_connections=200"]
```

After this change a 5-VU smoke scored 585/585 checks; the 100-VU run no longer hit the 10-second timeout wall.

## Takeaway

Three things to carry forward:

1. **Never trust a pool default.** Every connection-pool library ships with a default that exists so the library starts up — not so it survives load. pgx is 4, `database/sql` is unlimited (which is its own footgun), Postgres is 100. Set them explicitly and document the calculation: per-service pool × number of services + headroom ≤ Postgres `max_connections`.
2. **When two endpoints behave differently under the same load, the difference is the bug.** "Catalog runs short SELECTs; cart runs multi-statement transactions" should have pointed at pool occupancy fifteen minutes earlier than it did. Don't reach for the prettier explanation (caching) when the duller one (transaction duration) is right in the code.
3. **Two bugs colliding produces a confusing symptom.** The "401 unauthorized" log lines made this look like an auth bug; pool starvation was the actual cause, but it manifested through the cookie path because signup itself was being starved. When the data is contradictory — "auth works in isolation, fails under load, and the failure surfaces as 401" — look for two independent bugs whose symptoms overlap.

## Related

- [k6-cookie-jar.md](./k6-cookie-jar.md) — the cookie-cascade bug that piggybacked on this one.
- [postgres-conn-leaks.md](./postgres-conn-leaks.md) — why `max_connections=200` was the right number, not 100.
- [ADR 6 — catalog reads, cache-aside, cursor pagination](../adr/006-catalog-reads-cache-aside-cursor-pagination.md) — the cache-aside design that *will*, once implemented, change the catalog side of the asymmetry above.
