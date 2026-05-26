import { check } from 'k6';
import { httpPost, httpGet, idempotencyHeaders } from './config.js';
import { authHeaders } from './auth.js';

// placeOrder posts to /checkout with an Idempotency-Key header and no body.
// The gateway reads items from the user's current cart. On success the
// gateway clears the cart (best-effort).
//
// Returns { ok: bool, orderId: string|null, status: int }.
export function placeOrder() {
  const res = httpPost('/checkout', null, {
    headers: Object.assign({}, idempotencyHeaders(), authHeaders()),
  });
  const ok = check(res, { 'checkout 201 or 409': (r) => r.status === 201 || r.status === 409 });
  if (!ok || res.status !== 201) {
    return { ok: false, orderId: null, status: res.status };
  }
  const body = res.json();
  return { ok: true, orderId: body.id, status: res.status };
}

export function getOrder(id) {
  const res = httpGet(`/orders/${id}`, { headers: authHeaders() });
  check(res, { 'order 200': (r) => r.status === 200 });
  return res.status === 200 ? res.json() : null;
}
