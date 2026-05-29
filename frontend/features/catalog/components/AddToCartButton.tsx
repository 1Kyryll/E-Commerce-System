"use client";
import { Button } from "@/components/ui/button";
import { ShoppingCart } from "@/components/icons";
import { useCart } from "@/features/cart/context";
import type { Product } from "../types";

export function AddToCartButton({
  product,
  disabled,
}: {
  product: Product;
  disabled?: boolean;
}) {
  const cart = useCart();

  function onClick() {
    cart.add({
      product_id: product.id,
      quantity: 1,
      name: product.name,
      price: product.price,
    });
  }

  return (
    <Button onClick={onClick} disabled={disabled} className="w-full">
      <ShoppingCart className="h-4 w-4" />
      Add to cart
    </Button>
  );
}
