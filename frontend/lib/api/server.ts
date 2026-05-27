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
