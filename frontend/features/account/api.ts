import "server-only";
import { serverApi, unwrap } from "@/lib/api/server";
import type { components } from "@/lib/types";

export type User = components["schemas"]["User"];
export type Order = components["schemas"]["Order"];

export async function getMe(): Promise<User> {
  const api = serverApi();
  return await unwrap(await api.GET("/me"));
}

// No list-orders endpoint in the gateway yet; the OpenAPI spec (frontend/lib/types.ts)
// only exposes GET /orders/{id} — there is no GET /orders or GET /me/orders path.
// Until the gateway adds one, this returns an empty list and OrdersTable renders a
// friendly empty state. Do not call a path that isn't in `paths` — it won't typecheck.
export async function listMyOrders(): Promise<Order[]> {
  return [];
}

export async function getOrder(id: string): Promise<Order> {
  const api = serverApi();
  return await unwrap(
    await api.GET("/orders/{id}", { params: { path: { id } } }),
  );
}
