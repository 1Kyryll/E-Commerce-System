// Smoke: 1 VU × 30s. Touches every endpoint to prove the suite + app + LGTM
// stack are wired correctly. Run before any of the heavy scripts.

import { sleep } from 'k6';
import { ensureAuthenticated } from './lib/auth.js';
import { listProducts, pickRandomProductId, getProduct } from './lib/catalog.js';
import { getCart, addItem, clearCart } from './lib/cart.js';
import { placeOrder, getOrder } from './lib/checkout.js';

export const options = {
  scenarios: {
    smoke: {
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
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
    sleep(1);
    return;
  }
  const pid = page.products[0].id;
  getProduct(pid);

  addItem(pid, 1);
  getCart();

  const result = placeOrder();
  if (result.ok && result.orderId) {
    getOrder(result.orderId);
  }
  // Defensive: clear the cart in case checkout failed and left items.
  clearCart();

  sleep(1);
}
