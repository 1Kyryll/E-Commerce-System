"use client";
import { createContext, useContext, type ReactNode } from "react";
import Link from "next/link";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatMoney } from "@/lib/money";
import { cn } from "@/lib/cn";
import type { Product } from "../types";

const Ctx = createContext<Product | null>(null);
function useProduct() {
  const p = useContext(Ctx);
  if (!p) throw new Error("ProductCard.* must be used inside ProductCard");
  return p;
}

function Root({
  product,
  children,
  className,
}: {
  product: Product;
  children: ReactNode;
  className?: string;
}) {
  return (
    <Ctx.Provider value={product}>
      <Card className={cn("flex flex-col h-full", className)}>{children}</Card>
    </Ctx.Provider>
  );
}

function Image({ className }: { className?: string }) {
  const p = useProduct();
  return (
    <div
      className={cn(
        "aspect-square bg-sub-accent-1/20 overflow-hidden",
        className,
      )}
      aria-label={p.name}
    />
  );
}

function Title({ className }: { className?: string }) {
  const p = useProduct();
  return (
    <Link
      href={`/products/${p.id}`}
      className={cn("font-medium text-secondary hover:underline", className)}
    >
      {p.name}
    </Link>
  );
}

function Price({ className }: { className?: string }) {
  const p = useProduct();
  return (
    <p className={cn("font-secondary text-secondary", className)}>
      {formatMoney(p.price)}
    </p>
  );
}

function StockBadge() {
  const p = useProduct();
  const inv = p.inventory_available;
  if (inv <= 0) return <Badge tone="danger">Out of stock</Badge>;
  if (inv < 5) return <Badge tone="accent">Only {inv} left</Badge>;
  return null;
}

function Body({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <Card.Body className={cn("flex flex-col gap-2 flex-1", className)}>
      {children}
    </Card.Body>
  );
}

function Footer({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return <Card.Footer className={className}>{children}</Card.Footer>;
}

export const ProductCard = Object.assign(Root, {
  Image,
  Title,
  Price,
  StockBadge,
  Body,
  Footer,
});
