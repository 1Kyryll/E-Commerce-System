"use client";
import { useState, type ReactNode } from "react";
import Link from "next/link";
import { Sheet } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Trash2, Plus, Minus, ShoppingCart } from "@/components/icons";
import { formatMoney } from "@/lib/money";
import { useCart } from "../context";
import type { components } from "@/lib/types";

type Money = components["schemas"]["Money"];

function Root({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  return (
    <Sheet open={open} onOpenChange={setOpen}>
      {children}
    </Sheet>
  );
}

function Trigger() {
  const { totalCount } = useCart();
  return (
    <Sheet.Trigger asChild>
      <Button intent="ghost" size="icon" aria-label="Open cart">
        <span className="relative">
          <ShoppingCart className="h-5 w-5" />
          {totalCount > 0 && (
            <span className="absolute -top-2 -right-2">
              <Badge tone="success">{totalCount}</Badge>
            </span>
          )}
        </span>
      </Button>
    </Sheet.Trigger>
  );
}

function Drawer({ children }: { children?: ReactNode }) {
  return (
    <Sheet.Content side="right">
      <Sheet.Header>
        <Sheet.Title>Your cart</Sheet.Title>
      </Sheet.Header>
      <Items />
      <Summary />
      {children}
    </Sheet.Content>
  );
}

function shortId(id: string): string {
  return id.length > 12 ? `${id.slice(0, 8)}…` : id;
}

function Items() {
  const { items, remove, setQty } = useCart();
  if (items.length === 0) {
    return (
      <Sheet.Body>
        <p className="text-sub-accent-1 text-center py-12">Your cart is empty.</p>
      </Sheet.Body>
    );
  }
  return (
    <Sheet.Body>
      <ul className="flex flex-col gap-4">
        {items.map((it) => (
          <li key={it.product_id} className="flex gap-3">
            <div className="flex-1">
              <p className="font-medium text-secondary">
                {it.name ?? (
                  <span className="font-mono text-sm">{shortId(it.product_id)}</span>
                )}
              </p>
              {it.price && (
                <p className="text-sm text-sub-accent-1">{formatMoney(it.price)}</p>
              )}
              <div className="flex items-center gap-2 mt-2">
                <Button
                  intent="ghost"
                  size="icon"
                  aria-label="Decrease quantity"
                  onClick={() => setQty(it.product_id, (it.quantity ?? 1) - 1)}
                >
                  <Minus className="h-4 w-4" />
                </Button>
                <span className="w-6 text-center">{it.quantity ?? 0}</span>
                <Button
                  intent="ghost"
                  size="icon"
                  aria-label="Increase quantity"
                  onClick={() => setQty(it.product_id, (it.quantity ?? 0) + 1)}
                >
                  <Plus className="h-4 w-4" />
                </Button>
                <Button
                  intent="ghost"
                  size="icon"
                  aria-label="Remove"
                  className="ml-auto"
                  onClick={() => remove(it.product_id)}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </li>
        ))}
      </ul>
    </Sheet.Body>
  );
}

function Summary() {
  const { items } = useCart();
  if (items.length === 0) return null;

  const priced = items.filter((it) => it.price);
  const currency =
    (priced[0]?.price as Money | undefined)?.currency ?? "USD";
  const subtotal = priced.reduce(
    (acc, it) => acc + Number((it.price as Money).amount) * (it.quantity ?? 0),
    0,
  );
  const allPriced = priced.length === items.length;

  return (
    <Sheet.Footer>
      <div className="flex items-center justify-between mb-3">
        <span className="text-secondary">Subtotal</span>
        <span className="font-secondary text-secondary">
          {allPriced
            ? formatMoney({ amount: subtotal.toFixed(4), currency })
            : "—"}
        </span>
      </div>
      <Sheet.Close asChild>
        <Button asChild className="w-full">
          <Link href="/checkout">Checkout</Link>
        </Button>
      </Sheet.Close>
    </Sheet.Footer>
  );
}

export const Cart = Object.assign(Root, { Trigger, Drawer, Items, Summary });
