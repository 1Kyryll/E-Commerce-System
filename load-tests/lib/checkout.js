import { check } from 'k6';
import { httpPost, httpGet, idempotencyHeaders } from './config.js';

// placeOrder posts to /checkout with an Idempotency-Key header and no body.
// The gateway reads items from the user's current cart. On success the
// gateway clears the cart (best-effort).
//
// Returns { ok: bool, orderId: string|null, status: int }.
export function placeOrder() {
  const res = httpPost('/checkout', null, { headers: idempotencyHeaders() });
  const ok = check(res, { 'checkout 201': (r) => r.status === 201 });
  if (!ok) return { ok: false, orderId: null, status: res.status };
  const body = res.json();
  return { ok: true, orderId: body.id, status: res.status };
}

export function getOrder(id) {
  const res = httpGet(`/orders/${id}`);
  check(res, { 'order 200': (r) => r.status === 200 });
  return res.status === 200 ? res.json() : null;
}
