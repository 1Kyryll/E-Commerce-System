// browse.js — browse-heavy profile. A VU lists products, pages through 1–3
// pages, opens 1–2 random product detail pages, then sleeps like a real user
// reading the page.
//
// Profile size is controlled via env vars:
//   K6_VUS       — target VU count (default 100)
//   K6_DURATION  — steady-state duration (default 3m)
// The ramp-up and ramp-down are fixed (1m / 30s) so changing K6_DURATION
// only changes the steady plateau.

import { sleep } from 'k6';
import { ensureAuthenticated } from './lib/auth.js';
import { listProducts, getProduct } from './lib/catalog.js';
import { randomFromArray } from './lib/config.js';

const TARGET_VUS = parseInt(__ENV.K6_VUS || '100', 10);
const STEADY = __ENV.K6_DURATION || '3m';

export const options = {
  scenarios: {
    browse: {
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

  // First page.
  let page = listProducts('', 20);
  sleep(1 + Math.random() * 2);

  // Scroll: 1–3 additional pages, following the cursor.
  const moreScrolls = 1 + Math.floor(Math.random() * 3);
  for (let i = 0; i < moreScrolls && page && page.next_page_cursor; i++) {
    page = listProducts(page.next_page_cursor, 20);
    sleep(1 + Math.random() * 2);
  }

  // Open 1–2 random product detail pages.
  if (page && page.products && page.products.length > 0) {
    const opens = 1 + Math.floor(Math.random() * 2);
    for (let i = 0; i < opens; i++) {
      const p = randomFromArray(page.products);
      if (p) {
        getProduct(p.id);
        sleep(2 + Math.random() * 3);
      }
    }
  }
}
