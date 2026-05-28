"use client";
import { useTransition } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ShoppingCart } from "@/components/icons";
import { Spinner } from "@/components/ui/spinner";

type AddItemResult = { ok: boolean; formError?: string };
type CartActionsModule = {
  addItemAction?: (input: {
    productId: string;
    quantity: number;
  }) => Promise<AddItemResult>;
};

export function AddToCartButton({
  productId,
  disabled,
}: {
  productId: string;
  disabled?: boolean;
}) {
  const [pending, start] = useTransition();
  function onClick() {
    start(async () => {
      try {
        // Cart actions module is implemented in Phase 5. Until then this
        // dynamic import resolves to null and we surface a friendly toast.
        const specifier = "@/features/cart/actions" as string;
        const mod = (await import(/* webpackIgnore: true */ specifier).catch(
          () => null,
        )) as CartActionsModule | null;
        if (!mod || typeof mod.addItemAction !== "function") {
          toast.error("Cart is not available yet.");
          return;
        }
        const res = await mod.addItemAction({ productId, quantity: 1 });
        if (res.ok) toast.success("Added to cart");
        else toast.error(res.formError ?? "Could not add to cart");
      } catch {
        toast.error("Could not add to cart");
      }
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
