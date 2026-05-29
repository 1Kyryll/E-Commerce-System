"use client";
import { useTransition } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ShoppingCart } from "@/components/icons";
import { Spinner } from "@/components/ui/spinner";
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
  const [pending, start] = useTransition();

  function onClick() {
    start(() => {
      cart.add({
        product_id: product.id,
        quantity: 1,
        name: product.name,
        price: product.price,
      });
      toast.success("Added to cart");
    });
  }

  return (
    <Button
      onClick={onClick}
      disabled={disabled || pending}
      className="w-full"
    >
      {pending ? <Spinner /> : <ShoppingCart className="h-4 w-4" />}
      Add to cart
    </Button>
  );
}
