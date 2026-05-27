import createClient from "openapi-fetch";
import type { paths } from "@/lib/types";

const BASE_URL = process.env.NEXT_PUBLIC_GATEWAY_URL ?? "/api";

export const browserApi = createClient<paths>({
  baseUrl: BASE_URL,
  credentials: "include",
});
