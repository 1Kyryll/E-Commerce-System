import { z } from "zod";

// Backend `/checkout` takes no request body — it pulls items from the
// authenticated user's cart at request time. The only required input is
// an `Idempotency-Key` header (a UUID the client generates per attempt)
// per docs/adr/003-idempotency-on-order-creation.md.
export const checkoutSchema = z.object({
  idempotency_key: z.string().uuid(),
});
export type CheckoutInput = z.infer<typeof checkoutSchema>;
