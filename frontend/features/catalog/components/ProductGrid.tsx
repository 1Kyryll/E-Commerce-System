import { ProductCard } from "./ProductCard";
import { AddToCartButton } from "./AddToCartButton";
import type { Product } from "../types";

export function ProductGrid({ items }: { items: Product[] }) {
  if (items.length === 0) {
    return (
      <p className="text-sub-accent-1 py-12 text-center">No products yet.</p>
    );
  }
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      {items.map((p) => (
        <ProductCard key={p.id} product={p}>
          <ProductCard.Image />
          <ProductCard.Body>
            <ProductCard.Title />
            <ProductCard.Price />
            <ProductCard.StockBadge />
          </ProductCard.Body>
          <ProductCard.Footer>
            <AddToCartButton
              productId={p.id}
              disabled={p.inventory_available <= 0}
            />
          </ProductCard.Footer>
        </ProductCard>
      ))}
    </div>
  );
}
