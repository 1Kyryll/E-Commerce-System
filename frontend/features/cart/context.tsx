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
  updateQtyAction,
} from "./actions";
import type { Cart, CartItem } from "./api";

type State = {
  items: CartItem[];
  pending: boolean;
  add: (item: CartItem) => void;
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
  | { type: "add"; item: CartItem }
  | { type: "remove"; productId: string }
  | { type: "setQty"; productId: string; quantity: number };

function reducer(state: CartItem[], patch: Patch): CartItem[] {
  switch (patch.type) {
    case "add": {
      const i = state.findIndex((x) => x.product_id === patch.item.product_id);
      if (i >= 0) {
        const next = [...state];
        next[i] = {
          ...next[i],
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
  initial: Cart;
  children: ReactNode;
}) {
  const [optimistic, applyOptimistic] = useOptimistic(
    (initial.items ?? []) as CartItem[],
    reducer,
  );
  const [pending, start] = useTransition();

  function add(item: CartItem) {
    start(async () => {
      applyOptimistic({ type: "add", item });
      const res = await addItemAction({
        productId: item.product_id,
        quantity: item.quantity ?? 1,
      });
      if (!res.ok) toast.error(res.formError);
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
    start(async () => {
      applyOptimistic({ type: "setQty", productId, quantity });
      const res = await updateQtyAction({ productId, quantity });
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
