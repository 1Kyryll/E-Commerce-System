// checkout.js — full purchase profile. A VU browses, fills its cart with
// 1–2 products, posts /checkout, and follows up with GET /orders/{id} to
// model a confirmation page view. ~20% of attempts will fail with payment
// declined (the FakeClient's DefaultDeclineRate); the script does not retry.

import { check, sleep } from 'k6';
import { ensureAuthenticated } from './lib/auth.js';
import { listProducts } from './lib/catalog.js';
import { addItem, clearCart } from './lib/cart.js';
import { placeOrder, getOrder } from './lib/checkout.js';
import { randomFromArray } from './lib/config.js';

const TARGET_VUS = parseInt(__ENV.K6_VUS || '100', 10);
const STEADY = __ENV.K6_DURATION || '3m';

export const options = {
  scenarios: {
    checkout: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: TARGET_VUS },
        { duration: STEADY, target: TARGET_VUS },
        { duration: '30s', target: 0 },
      ],
      gracefulRampDown: '15s',
    },
  },
};

const RUN_ID = `${Date.now()}`;
export function setup() {
  return { runId: RUN_ID };
}

export default function (data) {
  ensureAuthenticated(data.runId);

  const page = listProducts('', 20);
  if (!page || !page.products || page.products.length === 0) {
    sleep(2);
    return;
  }
  sleep(0.5 + Math.random() * 1.5);

  // Add 1–2 distinct products.
  const adds = 1 + Math.floor(Math.random() * 2);
  const seen = new Set();
  for (let i = 0; i < adds; i++) {
    const p = randomFromArray(page.products);
    if (!p || seen.has(p.id)) continue;
    seen.add(p.id);
    addItem(p.id, 1);
    sleep(0.3 + Math.random() * 0.8);
  }

  // Checkout.
  const result = placeOrder();
  check(result, { 'checkout ok or 409': (r) => r.ok || r.status === 409 });
  if (result.ok && result.orderId) {
    sleep(1);
    getOrder(result.orderId);
  } else {
    // Cart wasn't cleared (checkout failed). Reset for next iteration.
    clearCart();
  }
  sleep(2 + Math.random() * 3);
}
