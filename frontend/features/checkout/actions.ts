"use server";
import { revalidatePath } from "next/cache";
import { serverApi, unwrap } from "@/lib/api/server";
import { userMessageFor } from "@/lib/api/errors";
import { checkoutSchema, type CheckoutInput } from "./schemas";

export type CheckoutResult =
  | { ok: true; orderId: string }
  | { ok: false; formError: string };

// Place an order from the caller's current cart. The backend pulls items
// from the session-scoped cart, so the only thing we send is the
// `Idempotency-Key` header. On success the cart is cleared server-side,
// so we revalidate the layout to drop the cart count from the nav.
export async function placeOrderAction(
  input: CheckoutInput,
): Promise<CheckoutResult> {
  const parsed = checkoutSchema.safeParse(input);
  if (!parsed.success) {
    return {
      ok: false,
      formError: parsed.error.issues[0]?.message ?? "Invalid input",
    };
  }
  try {
    const api = serverApi();
    const order = await unwrap(
      await api.POST("/checkout", {
        params: { header: { "Idempotency-Key": parsed.data.idempotency_key } },
      }),
    );
    revalidatePath("/", "layout");
    return { ok: true, orderId: order.id };
  } catch (err) {
    return { ok: false, formError: userMessageFor(err) };
  }
}
