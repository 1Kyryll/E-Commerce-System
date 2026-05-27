# Postgres `too many clients already` from leaked service restarts

## Symptom

After raising the pgx pool to 25 per service (see [pgx-pool-defaults.md](./pgx-pool-defaults.md)) and bumping Postgres `max_connections` from 100 to 200, an unexpected error started showing up mid-debugging:

```
psql: error: connection to server on socket "/var/run/postgresql/.s.PGSQL.5432"
failed: FATAL: sorry, too many clients already
```

`docker logs ecom-postgres` confirmed it server-side:

```
2026-05-25 20:59:31 UTC [11109] LOG:  connection received: host=[local]
2026-05-25 20:59:31 UTC [11109] FATAL: sorry, too many clients already
2026-05-25 20:59:33 UTC [11122] LOG:  connection received: host=[local]
2026-05-25 20:59:33 UTC [11122] FATAL: sorry, too many clients already
```

Strange, because the arithmetic should have worked: 4 services × 25 conns/pool = 100 steady-state connections, with 100 more reserved for `psql` sessions, monitoring, and one-off scripts. The error fired despite only four services running and no obvious source of extra load. Postgres was reporting itself out of slots even though our own services couldn't be holding more than 100 between them.

## First hypothesis

The first guess was that one of the services was opening more connections than its pool size suggested — that the pgx default kicks in somewhere we missed, or that some code path bypasses the pool. That theory died on a quick check: `netstat -ano | grep :9001` showed exactly one PID for each service, so no duplicated processes were running, and reading `internal/database/pool.go` confirmed that every binary goes through the same `NewPool` entry point that now defaults to 25.

The second guess was that the docker-compose change to `max_connections=200` hadn't actually taken effect — that we'd `docker compose restart`-ed (preserves command) instead of `docker compose up -d`-ing (applies new command). `docker inspect ecom-postgres --format '{{.Config.Cmd}}'` showed `[postgres -c max_connections=200]` and `SHOW max_connections;` returned `200`, so that ruled out the config too. Postgres was genuinely at its cap.

## Diagnosis

The decisive query was the one we couldn't run because we couldn't connect. Once Postgres was at the cap, *any* new client (including `psql`) bounced — including the `psql` session that would have told us who was holding the connections. The way around that is to query `pg_stat_activity` from a process that still has an open session, or to bounce Postgres briefly and snapshot immediately after restart.

We took the bounce path. `docker restart ecom-postgres` (~2 seconds), then `SELECT count(*) FROM pg_stat_activity;` returned **6**, climbing slowly as the still-running Go services reopened their pools. The data was intact — named volumes persist across container restarts. The container came back up running the same `postgres -c max_connections=200` command, so nothing about the configuration had changed. What changed was the connection state: every TCP socket Postgres held was dropped, and only the *currently running* services reconnected.

That gave the answer. Across the debugging session we had killed and restarted each Go service many times — `Ctrl+C` in the terminal running `make run-cart`, then `make run-cart` again, then a different test, then restart again. Each abrupt termination of a Go process closed its end of the TCP connections to Postgres, but Postgres's end only learns a peer is gone when the TCP keepalive interval expires (default ~2 hours on Linux without explicit `tcp_keepalives_*` tuning). Until then, those connections sit in `pg_stat_activity` as `idle` or `idle in transaction`, holding slots against `max_connections` without doing any work.

At 25 conns × maybe 8 service restarts during the debugging session = 200 leaked connections waiting to be reclaimed. Combined with the live services' real pools and a `psql` session or two, Postgres ran out of room.

## Root cause

When a pgx-using process is killed (not gracefully shut down), Postgres doesn't immediately notice the client is gone. Each killed process leaves its pool connections sitting in `pg_stat_activity` until Postgres's TCP keepalive declares them dead, which by default takes hours. Iterative debugging that restarts services repeatedly accumulates these leaked slots until `max_connections` is exhausted — independent of whether the currently running services are well-behaved.

## Fix

For the immediate cap-exhaustion: `docker restart ecom-postgres`. The named volume keeps the data; the container reset wipes every TCP socket on the server side. Services that are still running will reconnect on their next query. This was the right tool here because the leak was already there and needed clearing.

For preventing the same situation during future debugging, two things help:

1. **Graceful shutdowns.** Every `cmd/*/main.go` already wires `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` and `defer pool.Close()`. When the terminal forwards `Ctrl+C` as SIGINT, pgx closes its connections cleanly and Postgres reclaims the slots immediately. The trap is `Ctrl+Break` or closing the terminal window outright on Windows — those don't deliver SIGINT, and pgx never runs the close path.
2. **Shorter Postgres TCP keepalive** if leaks become routine. Postgres exposes `tcp_keepalives_idle`, `tcp_keepalives_interval`, `tcp_keepalives_count`. Setting `tcp_keepalives_idle=30` and `tcp_keepalives_interval=10` makes Postgres notice a dead client within ~30 seconds rather than hours. Not done yet — the bounce-on-demand approach is fine for a one-developer learning project, but it's the right knob if this codebase ever runs in CI where many service generations may overlap.

A diagnostic query worth keeping in a scratchpad for next time:

```sql
SELECT state, count(*), max(now() - state_change) AS oldest
  FROM pg_stat_activity
 WHERE datname='ecommerce'
 GROUP BY state
 ORDER BY count(*) DESC;
```

A large number of rows with `state='idle'` and an `oldest` age in the tens of minutes is the leak fingerprint.

## Takeaway

Three things to keep:

1. **Connection pools leak across process restarts unless you trap the kill signal.** This is true for every connection-pool library, every database, every language. A clean `pool.Close()` on shutdown is not a "good citizen" detail — it's the difference between development being smooth and a developer thinking the database is broken.
2. **When you can't query the database to find out who's connected, bounce it.** Named volumes mean no data loss; the inability to introspect a maxed-out Postgres is the kind of problem where the fix has to come before the diagnosis. Snapshot `pg_stat_activity` immediately after the restart if you want to know who reconnects fastest — that's often a hint about which service is doing the most work.
3. **`max_connections=200` looked like overkill until it wasn't.** The 2× headroom over the math wasn't waste — it absorbed exactly the kind of accumulated leak that happens during debugging. Cap sizing should plan for one full generation of leaked connections, not just the live ones.

## Related

- [pgx-pool-defaults.md](./pgx-pool-defaults.md) — why the per-service pool is 25; the calculation that decided 200 was a safe ceiling.
- [k6-cookie-jar.md](./k6-cookie-jar.md) — same debugging session; restarts to test fixes are what fed this leak.
