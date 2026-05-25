import { ensureAuthenticated } from './lib/auth.js';
import { addItem, getCart } from './lib/cart.js';
import { httpGet } from './lib/config.js';

export const options = { vus: 1, iterations: 1 };

export default function () {
    const runId = `dbg_${Date.now()}`;
    ensureAuthenticated(runId);

    // Dump cookies after signup.
    const jar = http_cookiejar();
    console.log('jar after signup:', JSON.stringify(jar));

    // Touch /cart and print everything.
    const cartRes = httpGet('/cart');
    console.log('GET /cart →', cartRes.status, cartRes.body);

    // Try /me too — same auth requirement, simpler.
    const meRes = httpGet('/me');
    console.log('GET /me →', meRes.status, meRes.body);
}

// k6 exposes the jar via http.cookieJar(); thin wrapper for the log line.
import http from 'k6/http';
function http_cookiejar() {
    return http.cookieJar().cookiesForURL('http://localhost:8080/');
}