import type { Metadata } from "next";
import { ThemeProvider } from "next-themes";
import { Toaster } from "sonner";
import { Suspense } from "react";
import { Nav } from "@/components/nav";
import { CartProvider, type EnrichedCartItem } from "@/features/cart/context";
import { getCart } from "@/features/cart/api";
import { getProduct } from "@/features/catalog/api";
import "./index.css";

export const metadata: Metadata = {
  title: "Storefront",
  description: "An e-commerce demo",
};

async function loadInitialCart(): Promise<{ items: EnrichedCartItem[] }> {
  try {
    const cart = await getCart();
    const items = await Promise.all(
      cart.items.map(async (it): Promise<EnrichedCartItem> => {
        try {
          const product = await getProduct(it.product_id);
          return { ...it, name: product.name, price: product.price };
        } catch {
          return it;
        }
      }),
    );
    return { items };
  } catch {
    return { items: [] };
  }
}

export default async function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const initialCart = await loadInitialCart();
  return (
    <html lang="en" suppressHydrationWarning className="h-full antialiased">
      <head>
        {/* Self-hosted font CSS lives in public/css and references its own
            woff2 files; loaded via <link> intentionally. next/font can't
            consume a pre-written @font-face stylesheet. */}
        {/* eslint-disable @next/next/no-css-tags */}
        <link rel="stylesheet" href="/css/satoshi.css" />
        <link rel="stylesheet" href="/css/general-sans.css" />
        {/* eslint-enable @next/next/no-css-tags */}
      </head>
      <body className="min-h-full flex flex-col bg-primary text-secondary font-primary">
        <ThemeProvider attribute="class" defaultTheme="light" enableSystem={false}>
          <CartProvider initial={initialCart}>
            <Suspense fallback={null}>
              <Nav />
            </Suspense>
            {children}
          </CartProvider>
          <Toaster richColors closeButton position="bottom-right" />
        </ThemeProvider>
      </body>
    </html>
  );
}
