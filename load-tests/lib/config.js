// Shared k6 helpers and runtime config. Imported by every script under
// load-tests/. Anything that varies between machines or runs lives here.

import http from 'k6/http';

// Gateway base URL. Override with BASE_URL env var; default matches local dev.
export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Default request params (timeout, etc.). k6's per-script options can still override.
export const REQUEST = {
  timeout: '10s',
};

// jsonHeaders returns the headers map for application/json bodies.
export function jsonHeaders() {
  return { 'Content-Type': 'application/json' };
}

// idempotencyHeaders returns headers including a fresh Idempotency-Key (UUID v4).
// k6 has no built-in UUID; use a 36-char random hex assembled in v4 layout.
export function idempotencyHeaders() {
  return Object.assign(jsonHeaders(), { 'Idempotency-Key': uuidv4() });
}

// uuidv4 generates a RFC4122 v4 UUID using crypto-strength randomness.
export function uuidv4() {
  // k6 exposes crypto via the WebCrypto-style globalThis.crypto in recent
  // builds. Fall back to Math.random() when absent (sufficient for load
  // testing where collisions are tolerated, not for production keys).
  const bytes = new Uint8Array(16);
  if (globalThis.crypto && globalThis.crypto.getRandomValues) {
    globalThis.crypto.getRandomValues(bytes);
  } else {
    for (let i = 0; i < 16; i++) bytes[i] = Math.floor(Math.random() * 256);
  }
  bytes[6] = (bytes[6] & 0x0f) | 0x40; // version 4
  bytes[8] = (bytes[8] & 0x3f) | 0x80; // variant 10
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

// randomFromArray returns one element. Returns undefined for empty input.
export function randomFromArray(arr) {
  if (!arr || arr.length === 0) return undefined;
  return arr[Math.floor(Math.random() * arr.length)];
}

// httpGet / httpPost / httpDelete are thin wrappers attaching default params.
// Centralized so we can add tracing headers or response checks in one place later.
export function httpGet(path, params = {}) {
  return http.get(`${BASE_URL}${path}`, Object.assign({}, REQUEST, params));
}
export function httpPost(path, body, params = {}) {
  return http.post(`${BASE_URL}${path}`, body, Object.assign({}, REQUEST, params));
}
export function httpDelete(path, params = {}) {
  return http.del(`${BASE_URL}${path}`, null, Object.assign({}, REQUEST, params));
}
