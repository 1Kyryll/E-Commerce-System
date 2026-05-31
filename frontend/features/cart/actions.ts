"use server";
import { revalidatePath } from "next/cache";
import { serverApi, unwrap } from "@/lib/api/server";
import { userMessageFor } from "@/lib/api/errors";
import { addItemSchema } from "./schemas";

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
    revalidatePath("/", "layout");
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
    revalidatePath("/", "layout");
    return { ok: true };
  } catch (err) {
    return { ok: false, formError: userMessageFor(err) };
  }
}

// Atomicity contract:
//   delta > 0           → single POST (additive)              ✓ atomic
//   newQuantity == 0    → single DELETE                       ✓ atomic
//   delta < 0, newQty>0 → DELETE then POST                    ⚠ NOT atomic
// The last case is the only non-atomic path. The backend does not
// currently expose a PATCH/PUT to set absolute quantity; until it does,
// shrinking a line without removing it has a small failure window where
// the line can end up cleared. Acceptable for MVP; document an issue
// to add a `PUT /cart/items/{product_id}` endpoint to fix it properly.
export async function adjustQtyAction(input: {
  productId: string;
  delta: number;
  previousQuantity: number;
}): Promise<ActionResult> {
  const { productId, delta, previousQuantity } = input;
  if (!Number.isInteger(delta) || delta === 0) return { ok: true };

  const newQuantity = previousQuantity + delta;
  if (newQuantity <= 0) return removeItemAction(productId);

  if (delta > 0) {
    return addItemAction({ productId, quantity: delta });
  }

  try {
    const api = serverApi();
    await unwrap(
      await api.DELETE("/cart/items/{product_id}", {
        params: { path: { product_id: productId } },
      }),
    );
    await unwrap(
      await api.POST("/cart/items", {
        body: { product_id: productId, quantity: newQuantity },
      }),
    );
    revalidatePath("/", "layout");
    return { ok: true };
  } catch (err) {
    return { ok: false, formError: userMessageFor(err) };
  }
}
