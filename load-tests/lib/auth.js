import { check, sleep } from 'k6';
import { httpPost, jsonHeaders } from './config.js';

// sessionToken is a module-scoped variable that PERSISTS across iterations of
// the same VU. We use it instead of k6's cookie jar because k6 clears the
// per-VU jar between iterations in this version — empirically verified. The
// JWT is stashed here once and attached via Cookie header on every
// authenticated call. authHeaders() below is the helper.
let sessionToken = '';

export function authHeaders() {
  return sessionToken ? { Cookie: `session=${sessionToken}` } : {};
}

// vuEmail returns a deterministic, unique email per (runId, VU). runId is
// the per-run nonce produced by setup(); __VU is k6's 1-indexed VU number.
// The scheme means same VU in two runs gets two different emails, so the
// unique-email constraint on users(email) doesn't bite between runs.
export function vuEmail(runId) {
  return `loadtest_${runId}_${__VU}@test.local`;
}

// PASSWORD is shared across all load-test users. Long enough to satisfy the
// gateway's 8-char minimum. Not a secret — these accounts are throwaway.
export const PASSWORD = 'loadtest-password';

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

// extractSession pulls the JWT out of a k6 response's parsed cookies.
function extractSession(res) {
  if (res && res.cookies && res.cookies.session && res.cookies.session.length > 0) {
    return res.cookies.session[0].value;
  }
  return '';
}

// ensureAuthenticated guarantees the VU has a sessionToken before returning.
// On the first iteration: sign up; if the email is already taken (409, can
// happen on overlapping runIds), fall back to login. Subsequent iterations
// short-circuit because sessionToken survives in module scope.
export function ensureAuthenticated(runId) {
  if (sessionToken) return;

  const email = vuEmail(runId);
  const signupRes = signup(email, PASSWORD);
  if (signupRes.status === 201) {
    sessionToken = extractSession(signupRes);
  } else if (signupRes.status === 409) {
    const loginRes = login(email, PASSWORD);
    if (loginRes.status === 200) {
      sessionToken = extractSession(loginRes);
    }
  }
  // Spread the auth burst when many VUs ramp up together.
  sleep(Math.random() * 0.5);
}
