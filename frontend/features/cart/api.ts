import "server-only";
import { serverApi, unwrap } from "@/lib/api/server";
import type { components } from "@/lib/types";

export type Cart = components["schemas"]["Cart"];
export type CartItem = components["schemas"]["CartItem"];

export async function getCart(): Promise<Cart> {
  const api = serverApi();
  const res = await api.GET("/cart", {
    next: { tags: ["cart"] },
  });
  return await unwrap(res);
}
