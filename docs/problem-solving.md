# Debugging and Problem Solving

While building this project I ran into a number of problems and bottlenecks that are worth documenting on their own. These notes may be a useful starting point for anyone reading or extending this codebase. The most common issues — how I diagnosed them and what the underlying causes turned out to be — live in [docs/issues/](./issues/).

## Issue format

Every file under `issues/` follows the same shape, so a reader can skim straight to the part they need:

- **Symptom** — what I actually saw (status codes, error messages, dashboard panels).
- **First hypothesis** — what I initially thought was going on, and why it turned out to be wrong (when it was).
- **Diagnosis** — the evidence that finally pointed at the cause: logs, SQL queries, k6 console output, diagnostic scripts.
- **Root cause** — one sentence on what was really happening.
- **Fix** — the change that resolved it, linked to the commit.
- **Takeaway** — what I'd do differently next time, or a sharper default to carry forward.

## Index

| Issue | Surface | Doc |
|---|---|---|
| k6's per-VU cookie jar clears between iterations | Load testing | [k6-cookie-jar.md](./issues/k6-cookie-jar.md) |
| pgx default pool of 4 connections starves under load | All services | [pgx-pool-defaults.md](./issues/pgx-pool-defaults.md) |
| Postgres "too many clients" from leaked connections | Database | [postgres-conn-leaks.md](./issues/postgres-conn-leaks.md) |
| OTel SDK schema-URL drift after upgrade | Observability | [otel-schema-url-drift.md](./issues/otel-schema-url-drift.md) |
| OTLP exporter silently uses an empty endpoint | Observability | [otlp-endpoint-empty.md](./issues/otlp-endpoint-empty.md) |
| `make` on Windows defaults to `cmd.exe`, breaking env-var recipes | Tooling | [make-windows-shell.md](./issues/make-windows-shell.md) |
| Depleted seed inventory hides as "checkout broken" | Test data | [test-data-inventory.md](./issues/test-data-inventory.md) |

## Patterns I keep seeing

Across the issues above, four themes recur. Each one is the kind of bug that a code review wouldn't catch — they live in defaults, env, and runtime, not in the source.

- **Defaults are rarely right under load.** pgx ships with a pool of 4; Postgres ships with 100 max connections; OTel SDK ships with no batching tuned for high cardinality. The defaults exist so the library starts up; they don't exist so it scales.
- **Empty env vars produce silent wrong behavior.** k6 falling back to `localhost:9090`, OTLP exporters dialing the empty string — neither raised a clear error. The first sign was always somewhere downstream.
- **Windows shell mismatch eats env-var prefixes.** GNU make on Windows, k6 invocations, WSL→host networking — all of them broke the same way until I forced `bash` explicitly.
- **Diagnostic logs beat hypothesis.** Every real cause in this project was found by printing the actual value, not by reasoning from first principles. The pattern: add one console.log or one SQL query, run the failing case, look at what the world actually said.

## What's not in here

Things deliberately not documented as "issues":

- Bugs I introduced and fixed within the same hour with no insight to extract. Those live in git history.
- Generic Go / Postgres / Docker knowledge widely available elsewhere. The index above is the issues this project specifically taught me.
- Design decisions and trade-offs — those belong under [docs/adr/](./adr/), not here.
