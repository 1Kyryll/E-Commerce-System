# Load tests (k6)

## Running with metrics → Grafana

```bash
K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9009/api/v1/push \
K6_PROMETHEUS_RW_PUSH_INTERVAL=5s \
k6 run --out experimental-prometheus-rw load-tests/checkout.js
```

Requires the LGTM stack from `observability/docker-compose.yml` to be running.
Metrics land in Mimir under `k6_*` series and are visible on the
`k6-load-test` Grafana dashboard.
