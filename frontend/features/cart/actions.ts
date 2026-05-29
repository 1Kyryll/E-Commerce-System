"use server";
import { revalidateTag } from "next/cache";
import { serverApi, unwrap } from "@/lib/api/server";
import { userMessageFor } from "@/lib/api/errors";
import { addItemSchema, updateQtySchema } from "./schemas";

export type ActionResult = { ok: true } | { ok: false; formError: string };

export async function addItemAction(input: {
  productId: string;
  quantity: number;
}): Promise<ActionResult> {
  const parsed = addItemSchema.safeParse(input);
  if (!parsed.success) {
    return {
      ok: false,
      formError: parsed.error.issues[0]?.message ?? "Invalid input",
    };
  }
  try {
    const api = serverApi();
    await unwrap(
      await api.POST("/cart/items", {
        body: {
          product_id: parsed.data.productId,
          quantity: parsed.data.quantity,
        },
      }),
    );
    revalidateTag("cart", "default");
    return { ok: true };
  } catch (err) {
    return { ok: false, formError: userMessageFor(err) };
  }
}

export async function removeItemAction(productId: string): Promise<ActionResult> {
  try {
    const api = serverApi();
    await unwrap(
      await api.DELETE("/cart/items/{product_id}", {
        params: { path: { product_id: productId } },
      }),
    );
    revalidateTag("cart", "default");
    return { ok: true };
  } catch (err) {
    return { ok: false, formError: userMessageFor(err) };
  }
}

// The API only exposes POST /cart/items (cumulative add) and
// DELETE /cart/items/{product_id}. There is no "set quantity" endpoint,
// so we emulate it: remove the line, then re-add at the desired quantity.
export async function updateQtyAction(input: {
  productId: string;
  quantity: number;
}): Promise<ActionResult> {
  const parsed = updateQtySchema.safeParse(input);
  if (!parsed.success) {
    return {
      ok: false,
      formError: parsed.error.issues[0]?.message ?? "Invalid input",
    };
  }
  if (parsed.data.quantity === 0) {
    return removeItemAction(parsed.data.productId);
  }
  try {
    const api = serverApi();
    // Remove existing line (no-op safe if it doesn't exist server-side: returns 200 with cart).
    await unwrap(
      await api.DELETE("/cart/items/{product_id}", {
        params: { path: { product_id: parsed.data.productId } },
      }),
    );
    await unwrap(
      await api.POST("/cart/items", {
        body: {
          product_id: parsed.data.productId,
          quantity: parsed.data.quantity,
        },
      }),
    );
    revalidateTag("cart", "default");
    return { ok: true };
  } catch (err) {
    return { ok: false, formError: userMessageFor(err) };
  }
}
