import type { components } from "@/lib/types";

export type Money = components["schemas"]["Money"];

export function formatMoney(m: Money): string {
  const major = Number(m.amount);
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: m.currency,
  }).format(major);
}
