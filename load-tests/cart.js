// cart.js — cart-manipulation profile. A VU browses, adds 1–3 items to its
// cart, occasionally views the cart, occasionally removes an item, then
// abandons (no checkout). Models the "deliberating shopper" cohort.

import { sleep } from 'k6';
import { ensureAuthenticated } from './lib/auth.js';
import { listProducts } from './lib/catalog.js';
import { getCart, addItem, removeItem } from './lib/cart.js';
import { randomFromArray } from './lib/config.js';

const TARGET_VUS = parseInt(__ENV.K6_VUS || '100', 10);
const STEADY = __ENV.K6_DURATION || '3m';

export const options = {
  scenarios: {
    cart: {
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
  sleep(1 + Math.random() * 2);

  // Pick 1–3 distinct products and add each.
  const adds = 1 + Math.floor(Math.random() * 3);
  const seen = new Set();
  const added = [];
  for (let i = 0; i < adds; i++) {
    const p = randomFromArray(page.products);
    if (!p || seen.has(p.id)) continue;
    seen.add(p.id);
    addItem(p.id, 1 + Math.floor(Math.random() * 2));
    added.push(p.id);
    sleep(0.5 + Math.random() * 1.5);
  }

  // 60% chance of viewing the cart.
  if (Math.random() < 0.6) {
    getCart();
    sleep(2 + Math.random() * 2);
  }

  // 25% chance of removing one added item.
  if (added.length > 0 && Math.random() < 0.25) {
    removeItem(added[Math.floor(Math.random() * added.length)]);
    sleep(1);
  }
  // Abandon — no checkout.
}
