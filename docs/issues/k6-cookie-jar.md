# k6's per-VU cookie jar is cleared between iterations

## Symptom

After the [pgx pool defaults](./pgx-pool-defaults.md) fix, signup was reliably returning 201 again — but cart and checkout endpoints still failed at around 50% in any sustained smoke run. The Grafana panel showed an unusual pattern: success rate started at ~100% for the first few seconds of a run, then dropped to roughly 50% and stayed there. That curve didn't fit a pool-starvation profile (which usually degrades smoothly with VU count); it fit a per-iteration mode change.

The actual failures clustered:
- A handful of `status=0` (k6 client timeouts left over from the previous issue, becoming rarer).
- A large number of `status=401 unauthorized` on every authenticated endpoint (`/cart`, `/cart/items`, `/checkout`, `/orders/*`), even though signup itself was now consistently succeeding.

So the question narrowed: signup returns 201 with a `Set-Cookie: session=...` header; cookie clearly arrives; yet the very next call from the same VU returns 401. Where does the cookie go between two HTTP calls in the same script?

## First hypothesis

The initial instinct was that the VUs were authenticating but losing the session — that whatever k6 was doing to track cookies wasn't behaving the way I expected. That turned out to be exactly right, but I burned time first chasing two wrong-feeling theories:

1. **Maybe signup was returning 201 without actually setting the cookie under load.** Ruled out by logging `signupRes.headers['Set-Cookie']` — the header was present and well-formed (`session=eyJ...; Path=/; Max-Age=86400; HttpOnly; SameSite=Lax`) on every successful signup.
2. **Maybe the gateway was rejecting the cookie because the JWT had expired or the secret had drifted.** Ruled out by decoding the JWT — it had a 24h expiry, the signature matched, and a manual `curl -H "Cookie: session=..."` against the same gateway instance returned 200.

What pointed at the real cause was logging the **state of k6's cookie jar at the start of each iteration**, before any HTTP call ran. The signup-on-iter-0 always succeeded and the post-signup jar contained the cookie. But the pre-iteration check on iter 1 reported the jar empty.

## Diagnosis

Adding two `console.log` lines to `lib/auth.js` (one before signup, one after) produced output like:

```
vu=1 iter=0 pre  has_session=false
vu=2 iter=0 pre  has_session=false
vu=1 iter=0 signup status=201 setcookie=session=eyJhbGc...
vu=1 iter=0 post has_session=true        ← cookie stored
vu=2 iter=0 signup status=201 setcookie=session=eyJhbGc...
vu=2 iter=0 post has_session=true
vu=1 iter=1 pre  has_session=false       ← cookie gone after one iteration
vu=2 iter=1 pre  has_session=false
vu=1 iter=1 signup status=409 body=email already exists
vu=1 iter=1 login  status=200            ← fallback login succeeds
vu=1 iter=1 post has_session=true        ← but jar empties again next iter
vu=2 iter=1 pre  has_session=false
...
```

That rules out everything except the jar itself. The cookie's `Max-Age` is 86400 (24h), the JWT's `exp` is also 24h out, and we never call `jar.clear()`. Yet the jar contents present at the end of iteration N are gone at the start of iteration N+1 for the same VU.

This contradicts what k6's docs lead you to expect — the per-VU cookie jar is described as persisting across iterations. In this k6 build it does not, at least for our combination of options (`experimental-prometheus-rw` output, `constant-vus` executor, vanilla `http` module). The behavior was confirmed three times: with `--vus 1 --duration 30s`, with `--vus 2 --duration 20s`, and with `--vus 5 --duration 20s`. Same result every time.

Tried as a workaround: explicitly calling `jar.set(url, name, value)` after each signup to force the cookie into the jar. Pre-iteration check on the next iteration still reported it gone. So even direct manipulation of the jar didn't survive the iteration boundary.

The 401s were the natural consequence: every iteration after the first started with an empty jar, k6 sent the cart/checkout request with no Cookie header, the gateway's auth middleware found no session, and returned 401.

## Root cause

k6's per-VU cookie jar, in the version we're running, is cleared between iterations rather than persisted. The Set-Cookie from `/auth/signup` is correctly received and stored — but the storage doesn't survive into the next iteration of the same VU, so every authenticated request after the first runs without credentials.

(This may be intentional behavior to model independent user sessions per iteration, or it may be a version-specific bug — I didn't dig into the k6 source to settle which. Either way the workaround is the same.)

## Fix

Sidestep the jar entirely. Stash the JWT in a **module-scoped variable** in `lib/auth.js` — module scope persists across iterations of the same VU, unlike the cookie jar — and attach it manually via a `Cookie` header on every authenticated call. Exposed as a small `authHeaders()` helper so individual scripts don't have to know the cookie name.

```javascript
// load-tests/lib/auth.js
let sessionToken = '';

export function authHeaders() {
  return sessionToken ? { Cookie: `session=${sessionToken}` } : {};
}

function extractSession(res) {
  if (res && res.cookies && res.cookies.session && res.cookies.session.length > 0) {
    return res.cookies.session[0].value;
  }
  return '';
}

export function ensureAuthenticated(runId) {
  if (sessionToken) return;
  const email = vuEmail(runId);
  const signupRes = signup(email, PASSWORD);
  if (signupRes.status === 201) {
    sessionToken = extractSession(signupRes);
  } else if (signupRes.status === 409) {
    const loginRes = login(email, PASSWORD);
    if (loginRes.status === 200) sessionToken = extractSession(loginRes);
  }
  sleep(Math.random() * 0.5);
}
```

Cart, checkout, and orders helpers each merge `authHeaders()` into their request headers:

```javascript
// load-tests/lib/cart.js
const res = httpPost(
  '/cart/items',
  JSON.stringify({ product_id: productId, quantity }),
  { headers: Object.assign({}, jsonHeaders(), authHeaders()) }
);
```

After this change, a 5-VU × 20s smoke scored **585/585 checks passing** — every cart and checkout call carries credentials across every iteration.

## Takeaway

Three things to keep in mind for the next time tooling surprises us:

1. **"Per-VU" is not the same as "persistent."** k6 hands you a per-VU cookie jar; that turned out to mean "scoped to the VU, but reset per iteration in our build." Documentation makes claims about persistence that didn't hold here. When a tool's behavior contradicts its docs, log the actual state — believe the logs, not the docs.
2. **Module scope is the reliable per-VU persistence in k6.** A `let foo = ''` at the top of a JS module survives across iterations of the same VU even when the cookie jar doesn't. That's the right escape hatch for any per-VU state load tests need to keep — JWTs, derived IDs, per-VU random seeds.
3. **One symptom (401) had two unrelated causes.** Pool starvation from [pgx-pool-defaults](./pgx-pool-defaults.md) caused early-run 401s; cookie-jar reset caused later-run 401s. Both got fixed independently and both fixes were necessary. The instinct to keep digging after the first fix only got a 7% → 50% improvement was the right one — partial fixes that don't restore 100% almost always mean a second bug is hiding underneath.

## Related

- [pgx-pool-defaults.md](./pgx-pool-defaults.md) — the connection-starvation issue that ran in parallel with this one and shared the 401 symptom.
- [postgres-conn-leaks.md](./postgres-conn-leaks.md) — same era of debugging, same load-test session.
