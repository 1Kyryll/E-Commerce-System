# E-Commerce Demo

See `docs/` for design. Layout follows `docs/code-architecture.md`.

## Quick start

```
make run-gateway     # backend gateway hello-world
make frontend-dev    # next dev server
```

## Observability

The LGTM stack runs from a dedicated compose:

```bash
docker compose -f observability/docker-compose.yml up -d
```

Then start app services with `OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317` in
`backend/.env`. Grafana: http://localhost:3000 (anonymous, admin).

### k6 metrics → Grafana

```bash
K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9009/api/v1/push \
K6_PROMETHEUS_RW_PUSH_INTERVAL=5s \
k6 run --out experimental-prometheus-rw load-tests/checkout.js
```

Metrics land in Mimir under `k6_*` series and show on the `k6-load-test` dashboard.
