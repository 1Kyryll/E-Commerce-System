import { check, sleep } from 'k6';
import { httpPost, jsonHeaders } from './config.js';

// vuEmail returns a deterministic, unique email per (runId, VU). runId is a
// per-run nonce produced by setup(); __VU is k6's 1-indexed VU number. The
// scheme means: same VU in two different runs gets two different emails, so
// the unique-email constraint on users(email) doesn't bite between runs.
export function vuEmail(runId) {
  return `loadtest_${runId}_${__VU}@test.local`;
}

// PASSWORD is shared across all load-test users. Long enough to satisfy the
// gateway's 8-char minimum. Not a secret — these accounts are throwaway.
export const PASSWORD = 'loadtest-password';

// signup posts to /auth/signup. The handler sets a session cookie that k6's
// per-VU cookie jar will reuse on subsequent calls automatically.
export function signup(email, password) {
  const res = httpPost(
    '/auth/signup',
    JSON.stringify({ name: email, email, password }),
    { headers: jsonHeaders() }
  );
  check(res, { 'signup 201': (r) => r.status === 201 });
  return res;
}

// login posts to /auth/login. Used as a fallback when signup returns 409
// (email already in use, e.g. from a previous run that overlapped runIds).
export function login(email, password) {
  const res = httpPost(
    '/auth/login',
    JSON.stringify({ email, password }),
    { headers: jsonHeaders() }
  );
  check(res, { 'login 200': (r) => r.status === 200 });
  return res;
}

// ensureAuthenticated runs once per VU on iteration 0. Subsequent iterations
// rely on the cookie jar. Falls back from signup → login on 409 collisions
// so re-running the same runId doesn't break.
export function ensureAuthenticated(runId) {
  // k6 starts iteration numbering at 0; __ITER is the current iteration.
  if (__ITER !== 0) return;
  const email = vuEmail(runId);
  const signupRes = signup(email, PASSWORD);
  if (signupRes.status === 409) {
    login(email, PASSWORD);
  }
  // Spread the auth burst across the ramp window by sleeping a small jitter
  // (0–500ms). Without this, every VU hits /auth/signup the moment it spawns.
  sleep(Math.random() * 0.5);
}
