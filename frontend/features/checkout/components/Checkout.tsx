"use client";
import { useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { formatMoney } from "@/lib/money";
import { useCart } from "@/features/cart/context";
import { placeOrderAction } from "../actions";
import type { components } from "@/lib/types";

type Money = components["schemas"]["Money"];

// Path B (single confirmation screen): the backend's /checkout endpoint
// takes no request body — items come from the server-side cart and the
// only client input is an Idempotency-Key UUID. We therefore skip the
// fake shipping/payment wizard and render a summary + Place Order button.
// The idempotency key is generated once per mount and reused for retries
// so a redelivered request hits the same parent on the server.
export function Checkout() {
  const router = useRouter();
  const { items } = useCart();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [idempotencyKey] = useState(() => crypto.randomUUID());

  const priced = items.filter((it) => it.price);
  const allPriced = priced.length === items.length && items.length > 0;
  const currency =
    (priced[0]?.price as Money | undefined)?.currency ?? "USD";
  const subtotal = useMemo(
    () =>
      priced.reduce(
        (acc, it) =>
          acc + Number((it.price as Money).amount) * (it.quantity ?? 0),
        0,
      ),
    [priced],
  );

  if (items.length === 0) {
    return (
      <Card>
        <Card.Body className="text-center py-12">
          <p className="text-sub-accent-1 mb-4">Your cart is empty.</p>
          <Button asChild>
            <Link href="/">Browse products</Link>
          </Button>
        </Card.Body>
      </Card>
    );
  }

  async function onPlaceOrder() {
    setSubmitting(true);
    setError(null);
    const res = await placeOrderAction({ idempotency_key: idempotencyKey });
    if (res.ok) {
      toast.success("Order placed");
      router.push(`/orders/${res.orderId}`);
      return;
    }
    setSubmitting(false);
    setError(res.formError);
    toast.error(res.formError);
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <Card className="lg:col-span-2">
        <Card.Header>
          <h2 className="font-secondary text-lg text-secondary">Review your order</h2>
        </Card.Header>
        <Card.Body>
          <ul className="flex flex-col gap-4">
            {items.map((it) => (
              <li
                key={it.product_id}
                className="flex items-center justify-between gap-4"
              >
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-secondary truncate">
                    {it.name ?? it.product_id}
                  </p>
                  {it.price && (
                    <p className="text-sm text-sub-accent-1">
                      {formatMoney(it.price)} × {it.quantity ?? 0}
                    </p>
                  )}
                </div>
                {it.price && (
                  <span className="font-secondary text-secondary">
                    {formatMoney({
                      amount: (
                        Number(it.price.amount) * (it.quantity ?? 0)
                      ).toFixed(4),
                      currency: it.price.currency,
                    })}
                  </span>
                )}
              </li>
            ))}
          </ul>
        </Card.Body>
      </Card>

      <Card>
        <Card.Header>
          <h2 className="font-secondary text-lg text-secondary">Summary</h2>
        </Card.Header>
        <Card.Body>
          <div className="flex items-center justify-between">
            <span className="text-secondary">Subtotal</span>
            <span className="font-secondary text-secondary">
              {allPriced
                ? formatMoney({ amount: subtotal.toFixed(4), currency })
                : "—"}
            </span>
          </div>
          {error && (
            <p className="text-sm text-red-600 mt-3" role="alert">
              {error}
            </p>
          )}
        </Card.Body>
        <Card.Footer>
          <Button
            className="w-full"
            onClick={onPlaceOrder}
            disabled={submitting}
          >
            {submitting ? (
              <>
                <Spinner className="h-4 w-4" /> Placing order…
              </>
            ) : (
              "Place order"
            )}
          </Button>
        </Card.Footer>
      </Card>
    </div>
  );
}
