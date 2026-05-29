import "server-only";
import { cookies, headers } from "next/headers";
import createClient from "openapi-fetch";
import type { paths } from "@/lib/types";
import { ApiError } from "./errors";

const BASE_URL = process.env.GATEWAY_URL ?? "http://localhost:8080";

export function serverApi() {
  return createClient<paths>({
    baseUrl: BASE_URL,
    fetch: async (req) => {
      const cookieHeader = (await cookies()).toString();
      const reqHeaders = new Headers(req.headers);
      if (cookieHeader) reqHeaders.set("cookie", cookieHeader);
      const traceparent = (await headers()).get("traceparent");
      if (traceparent) reqHeaders.set("traceparent", traceparent);
      return fetch(new Request(req, { headers: reqHeaders }));
    },
  });
}

export async function unwrap<T>(
  res: { data?: T; error?: unknown; response: Response },
): Promise<T> {
  if (res.data !== undefined) return res.data;
  const status = res.response.status;
  let code = "unknown";
  let message = res.response.statusText;
  if (res.error && typeof res.error === "object") {
    const e = res.error as { code?: string; message?: string };
    if (e.code) code = e.code;
    if (e.message) message = e.message;
  }
  throw new ApiError(status, code, message);
}

// The backend sets cookies (notably `session`) on its responses, but Next.js
// does NOT auto-forward Set-Cookie from a Server-Action's internal fetch to
// the browser. Call this after auth mutations to relay them.
//
// Must run inside a Server Action or Route Handler — cookies().set() throws
// in RSC reads.
export async function forwardSetCookies(res: Response): Promise<void> {
  const list =
    typeof res.headers.getSetCookie === "function"
      ? res.headers.getSetCookie()
      : [];
  if (list.length === 0) return;
  const store = await cookies();
  for (const raw of list) {
    const parsed = parseSetCookie(raw);
    if (!parsed) continue;
    store.set(parsed.name, parsed.value, parsed.opts);
  }
}

type ParsedSetCookie = {
  name: string;
  value: string;
  opts: {
    path?: string;
    domain?: string;
    maxAge?: number;
    expires?: Date;
    sameSite?: "lax" | "strict" | "none";
    secure?: boolean;
    httpOnly?: boolean;
  };
};

function parseSetCookie(s: string): ParsedSetCookie | null {
  const parts = s.split(";").map((p) => p.trim()).filter(Boolean);
  if (parts.length === 0) return null;
  const first = parts[0];
  const eq = first.indexOf("=");
  if (eq < 0) return null;
  const name = first.slice(0, eq).trim();
  const value = first.slice(eq + 1).trim();
  const opts: ParsedSetCookie["opts"] = {};
  for (let i = 1; i < parts.length; i++) {
    const attr = parts[i];
    const aeq = attr.indexOf("=");
    const k = (aeq < 0 ? attr : attr.slice(0, aeq)).toLowerCase().trim();
    const v = aeq < 0 ? "" : attr.slice(aeq + 1).trim();
    switch (k) {
      case "path":
        opts.path = v;
        break;
      case "domain":
        opts.domain = v;
        break;
      case "max-age": {
        const n = Number.parseInt(v, 10);
        if (Number.isFinite(n)) opts.maxAge = n;
        break;
      }
      case "expires": {
        const d = new Date(v);
        if (!Number.isNaN(d.getTime())) opts.expires = d;
        break;
      }
      case "samesite": {
        const lower = v.toLowerCase();
        if (lower === "lax" || lower === "strict" || lower === "none") {
          opts.sameSite = lower;
        }
        break;
      }
      case "secure":
        opts.secure = true;
        break;
      case "httponly":
        opts.httpOnly = true;
        break;
    }
  }
  return { name, value, opts };
}
