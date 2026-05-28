import { formatMoney } from "@/lib/money";
import { Badge } from "@/components/ui/badge";
import { Surface } from "@/components/ui/surface";
import { AddToCartButton } from "./AddToCartButton";
import type { Product } from "../types";

export function ProductDetail({ product }: { product: Product }) {
  const inv = product.inventory_available;
  const outOfStock = inv <= 0;
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
      <Surface
        rounded
        className="aspect-square bg-sub-accent-1/20 overflow-hidden"
        aria-label={product.name}
      />
      <div className="flex flex-col gap-4">
        <h1 className="text-3xl font-secondary font-semibold text-secondary">
          {product.name}
        </h1>
        <p className="text-xl font-secondary text-secondary">
          {formatMoney(product.price)}
        </p>
        {outOfStock ? (
          <Badge tone="danger">Out of stock</Badge>
        ) : inv < 5 ? (
          <Badge tone="accent">Only {inv} left</Badge>
        ) : null}
        <div className="pt-4">
          <AddToCartButton productId={product.id} disabled={outOfStock} />
        </div>
      </div>
    </div>
  );
}
