import { check } from 'k6';
import { httpGet, httpPost, httpDelete, jsonHeaders } from './config.js';

export function getCart() {
  const res = httpGet('/cart');
  check(res, { 'cart 200': (r) => r.status === 200 });
  return res.status === 200 ? res.json() : null;
}

export function addItem(productId, quantity) {
  const res = httpPost(
    '/cart/items',
    JSON.stringify({ product_id: productId, quantity }),
    { headers: jsonHeaders() }
  );
  check(res, { 'add item 200': (r) => r.status === 200 });
  return res;
}

export function removeItem(productId) {
  const res = httpDelete(`/cart/items/${productId}`);
  check(res, { 'remove item 200': (r) => r.status === 200 });
  return res;
}

export function clearCart() {
  const res = httpDelete('/cart');
  check(res, { 'clear cart 204': (r) => r.status === 204 });
  return res;
}
