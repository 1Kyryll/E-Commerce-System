import { check } from 'k6';
import { httpGet, httpPost, httpDelete, jsonHeaders } from './config.js';
import { authHeaders } from './auth.js';

export function getCart() {
  const res = httpGet('/cart', { headers: authHeaders() });
  check(res, { 'cart 200': (r) => r.status === 200 });
  return res.status === 200 ? res.json() : null;
}

export function addItem(productId, quantity) {
  const res = httpPost(
    '/cart/items',
    JSON.stringify({ product_id: productId, quantity }),
    { headers: Object.assign({}, jsonHeaders(), authHeaders()) }
  );
  check(res, { 'add item 200': (r) => r.status === 200 });
  return res;
}

export function removeItem(productId) {
  const res = httpDelete(`/cart/items/${productId}`, { headers: authHeaders() });
  check(res, { 'remove item 200': (r) => r.status === 200 });
  return res;
}

export function clearCart() {
  const res = httpDelete('/cart', { headers: authHeaders() });
  check(res, { 'clear cart 204': (r) => r.status === 204 });
  return res;
}
