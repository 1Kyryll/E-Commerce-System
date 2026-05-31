import Link from "next/link";
import { notFound } from "next/navigation";
import { Container } from "@/components/container";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Check } from "@/components/icons";
import { formatMoney } from "@/lib/money";
import { serverApi, unwrap } from "@/lib/api/server";
import { ApiError } from "@/lib/api/errors";
import type { components } from "@/lib/types";

type Order = components["schemas"]["Order"];

function statusTone(status: Order["status"]): "success" | "accent" | "danger" | "neutral" {
  switch (status) {
    case "paid":
      return "success";
    case "pending":
      return "accent";
    case "failed":
      return "danger";
    case "cancelled":
      return "neutral";
    default:
      return "neutral";
  }
}

export default async function OrderPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  let order: Order;
  try {
    const api = serverApi();
    order = await unwrap(
      await api.GET("/orders/{id}", { params: { path: { id } } }),
    );
  } catch (err) {
    if (err instanceof ApiError && err.isNotFound) notFound();
    throw err;
  }

  return (
    <Container className="py-10 flex-1 max-w-3xl">
      <div className="flex items-center gap-3 mb-6">
        <span className="h-10 w-10 rounded-full bg-action-primary text-primary flex items-center justify-center">
          <Check className="h-5 w-5" />
        </span>
        <div className="flex-1">
          <h1 className="text-2xl font-secondary text-secondary">Order placed</h1>
          <p className="text-sm text-sub-accent-1 font-mono">#{order.id}</p>
        </div>
        <Badge tone={statusTone(order.status)}>{order.status}</Badge>
      </div>

      <Card>
        <Card.Header>
          <h2 className="font-secondary text-secondary">Items</h2>
        </Card.Header>
        <Card.Body>
          <ul className="flex flex-col gap-4">
            {order.items.map((it) => (
              <li
                key={it.product_id}
                className="flex items-center justify-between gap-4"
              >
                <div className="flex-1 min-w-0">
                  <p className="font-mono text-sm text-secondary truncate">
                    {it.product_id}
                  </p>
                  <p className="text-sm text-sub-accent-1">
                    {formatMoney(it.unit_price)} × {it.quantity}
                  </p>
                </div>
                <span className="font-secondary text-secondary">
                  {formatMoney({
                    amount: (
                      Number(it.unit_price.amount) * it.quantity
                    ).toFixed(4),
                    currency: it.unit_price.currency,
                  })}
                </span>
              </li>
            ))}
          </ul>
        </Card.Body>
        <Card.Footer>
          <div className="flex items-center justify-between">
            <span className="text-secondary">Total</span>
            <span className="font-secondary text-secondary text-lg">
              {formatMoney(order.total)}
            </span>
          </div>
        </Card.Footer>
      </Card>

      {order.created_at && (
        <p className="text-sm text-sub-accent-1 mt-4">
          Placed on {new Date(order.created_at).toLocaleString()}
        </p>
      )}

      <div className="mt-6">
        <Button asChild intent="outline">
          <Link href="/">Continue shopping</Link>
        </Button>
      </div>
    </Container>
  );
}
