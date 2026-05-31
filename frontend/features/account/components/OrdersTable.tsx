import Link from "next/link";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatMoney } from "@/lib/money";
import type { Order } from "../api";

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

export function OrdersTable({ orders }: { orders: Order[] }) {
  if (orders.length === 0) {
    return (
      <Card>
        <Card.Body>
          <p className="text-sub-accent-1 text-center py-8">
            You haven&apos;t placed any orders yet.
          </p>
        </Card.Body>
      </Card>
    );
  }
  return (
    <Card>
      <ul className="divide-y divide-sub-accent-1/40">
        {orders.map((o) => (
          <li key={o.id} className="p-4 flex items-center justify-between gap-4">
            <div className="min-w-0">
              <Link
                href={`/orders/${o.id}`}
                className="font-medium text-secondary hover:underline font-mono text-sm"
              >
                #{o.id.slice(0, 8)}
              </Link>
              {o.created_at && (
                <p className="text-xs text-sub-accent-1">
                  {new Date(o.created_at).toLocaleString()}
                </p>
              )}
            </div>
            <div className="flex items-center gap-3">
              <Badge tone={statusTone(o.status)}>{o.status}</Badge>
              <span className="font-secondary text-secondary">
                {formatMoney(o.total)}
              </span>
            </div>
          </li>
        ))}
      </ul>
    </Card>
  );
}
