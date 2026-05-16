# Load Testing

One of the goals of this system, is to guarantee it is *Scalable, Durable and Available* under huge amount of traffic. Ensure that the data is consistent, and users are getting the best UX possible. **Grafana k6** framework is the most suitable and robust candidate to simulate real-world load and test user's application flow. 

## Why Grafana k6? 

**k6** itself is written in Go, scripted in JavaScript and the most important point is that it is made by *Grafana Labs*. Because k6 is made by *Grafana Labs*, it is native with **Grafana LGTM framework** which we will use as an observability solution here, k6 emits metrics in Prometheus format, so a load test populates the same dashboards as your real traffic. One of the best k6 features is that it that it allows us to write user flows, **not** endpoint hammers - a realistic e-commerce test isn't 5000 RPS to `/products`, it's "70% of users browse, 20% search, 8% add to cart, 2% check out, with 3–7 second think times between actions." Building that scenario teaches you behavioral modeling, which is a different and more valuable skill than maximizing requests per second.

**k6's thresholds are SLOs in disguise.** They are declared at the top of the script: p95 under 300ms, error rate under 1%, checkout success above 99%. The test fails if they're breached. That means a CI pipeline literally enforces SLOs on every deploy, which is exactly how staff-level SREs think about reliability. 

## Where scripts live?

They live in the `load-tests` directory, everything else is configured in the `observability` directory. Example load test script written in JavaScript might look like this: 

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const checkoutSuccess = new Rate('checkout_success');

export const options = {
  stages: [
    { duration: '1m', target: 50 },    // ramp up
    { duration: '3m', target: 50 },    // baseline
    { duration: '1m', target: 200 },   // peak
    { duration: '1m', target: 0 },     // ramp down
  ],
  thresholds: {
    'http_req_duration{endpoint:browse}':   ['p(95)<300'],
    'http_req_duration{endpoint:checkout}': ['p(95)<500'],
    http_req_failed:  ['rate<0.01'],
    checkout_success: ['rate>0.95'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // Most users browse
  const list = http.get(`${BASE_URL}/products`, { tags: { endpoint: 'browse' } });
  check(list, { 'list 200': r => r.status === 200 });

  // ~10% proceed to checkout
  if (Math.random() < 0.1) {
    const res = http.post(`${BASE_URL}/checkout`,
      JSON.stringify({ product_id: 'p_123' }),
      {
        headers: {
          'Content-Type':    'application/json',
          'Idempotency-Key': crypto.randomUUID(),
        },
        tags: { endpoint: 'checkout' },
      });
    checkoutSuccess.add(res.status === 201);
  }

  sleep(Math.random() * 4 + 3); // 3-7s think time
}
```
