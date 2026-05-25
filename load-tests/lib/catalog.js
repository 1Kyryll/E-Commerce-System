import { check } from 'k6';
import { httpGet, randomFromArray } from './config.js';

// listProducts hits GET /products with optional cursor + pageSize. Returns
// the parsed JSON body, or null on non-200.
export function listProducts(cursor = '', pageSize = 20) {
  const qs = [];
  if (cursor) qs.push(`cursor=${encodeURIComponent(cursor)}`);
  if (pageSize) qs.push(`page_size=${pageSize}`);
  const path = qs.length ? `/products?${qs.join('&')}` : '/products';
  const res = httpGet(path);
  const ok = check(res, { 'products 200': (r) => r.status === 200 });
  if (!ok) return null;
  return res.json();
}

// getProduct hits GET /products/{id}. Returns parsed product or null.
export function getProduct(id) {
  const res = httpGet(`/products/${id}`);
  const ok = check(res, { 'product 200': (r) => r.status === 200 });
  if (!ok) return null;
  return res.json();
}

// pickRandomProductId fetches one page of products and returns a random id
// from it. Useful for scripts that don't care which product they're working
// with. Returns null if the catalog is empty or the call failed.
export function pickRandomProductId() {
  const page = listProducts('', 20);
  if (!page || !page.products || page.products.length === 0) return null;
  return randomFromArray(page.products).id;
}
