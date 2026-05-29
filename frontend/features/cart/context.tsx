"use client";
import {
  createContext,
  useContext,
  useOptimistic,
  useTransition,
  type ReactNode,
} from "react";
import { toast } from "sonner";
import {
  addItemAction,
  removeItemAction,
  adjustQtyAction,
} from "./actions";
import type { CartItem } from "./api";
import type { components } from "@/lib/types";

type Money = components["schemas"]["Money"];

// Backend's CartItem is `{ product_id, quantity, added_at? }` only.
// We carry name + price client-side so the drawer renders product info
// without a per-line refetch. The fields are populated when items are
// added (by AddToCartButton, which knows the product) and when the
// layout enriches initial items via getProduct on first load.
export type EnrichedCartItem = CartItem & {
  name?: string;
  price?: Money;
};

type State = {
  items: EnrichedCartItem[];
  pending: boolean;
  add: (item: EnrichedCartItem) => void;
  remove: (productId: string) => void;
  setQty: (productId: string, quantity: number) => void;
  totalCount: number;
};

const Ctx = createContext<State | null>(null);

export function useCart() {
  const v = useContext(Ctx);
  if (!v) throw new Error("useCart must be used inside <CartProvider>");
  return v;
}

type Patch =
  | { type: "add"; item: EnrichedCartItem }
  | { type: "remove"; productId: string }
  | { type: "setQty"; productId: string; quantity: number };

function reducer(state: EnrichedCartItem[], patch: Patch): EnrichedCartItem[] {
  switch (patch.type) {
    case "add": {
      const i = state.findIndex((x) => x.product_id === patch.item.product_id);
      if (i >= 0) {
        const next = [...state];
        next[i] = {
          ...next[i],
          // Newer enrichment wins so adds from product pages overwrite stale
          // entries that lost their name/price after a refresh.
          name: patch.item.name ?? next[i].name,
          price: patch.item.price ?? next[i].price,
          quantity: (next[i].quantity ?? 0) + (patch.item.quantity ?? 1),
        };
        return next;
      }
      return [...state, patch.item];
    }
    case "remove":
      return state.filter((x) => x.product_id !== patch.productId);
    case "setQty":
      if (patch.quantity <= 0) {
        return state.filter((x) => x.product_id !== patch.productId);
      }
      return state.map((x) =>
        x.product_id === patch.productId ? { ...x, quantity: patch.quantity } : x,
      );
  }
}

export function CartProvider({
  initial,
  children,
}: {
  initial: { items: EnrichedCartItem[] };
  children: ReactNode;
}) {
  const [optimistic, applyOptimistic] = useOptimistic(
    initial.items,
    reducer,
  );
  const [pending, start] = useTransition();

  function add(item: EnrichedCartItem) {
    start(async () => {
      applyOptimistic({ type: "add", item });
      const res = await addItemAction({
        productId: item.product_id,
        quantity: item.quantity ?? 1,
      });
      if (res.ok) toast.success("Added to cart");
      else toast.error(res.formError);
    });
  }

  function remove(productId: string) {
    start(async () => {
      applyOptimistic({ type: "remove", productId });
      const res = await removeItemAction(productId);
      if (!res.ok) toast.error(res.formError);
    });
  }

  function setQty(productId: string, quantity: number) {
    const previousQuantity =
      optimistic.find((x) => x.product_id === productId)?.quantity ?? 0;
    const delta = quantity - previousQuantity;
    if (delta === 0) return;
    start(async () => {
      applyOptimistic({ type: "setQty", productId, quantity });
      const res = await adjustQtyAction({ productId, delta, previousQuantity });
      if (!res.ok) toast.error(res.formError);
    });
  }

  const totalCount = optimistic.reduce(
    (sum, it) => sum + (it.quantity ?? 0),
    0,
  );

  return (
    <Ctx.Provider
      value={{ items: optimistic, pending, add, remove, setQty, totalCount }}
    >
      {children}
    </Ctx.Provider>
  );
}
