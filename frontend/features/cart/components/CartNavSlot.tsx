"use client";
import { Cart } from "./Cart";

export function CartNavSlot() {
  return (
    <Cart>
      <Cart.Trigger />
      <Cart.Drawer />
    </Cart>
  );
}
