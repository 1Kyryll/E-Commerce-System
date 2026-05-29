import type { Metadata } from "next";
import { ThemeProvider } from "next-themes";
import { Toaster } from "sonner";
import { Suspense } from "react";
import { Nav } from "@/components/nav";
import { CartProvider } from "@/features/cart/context";
import { getCart, type Cart } from "@/features/cart/api";
import "./index.css";

export const metadata: Metadata = {
  title: "Storefront",
  description: "An e-commerce demo",
};

async function loadInitialCart(): Promise<Cart> {
  try {
    return await getCart();
  } catch {
    return {
      id: "00000000-0000-0000-0000-000000000000",
      user_id: "00000000-0000-0000-0000-000000000000",
      items: [],
    };
  }
}

export default async function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const initialCart = await loadInitialCart();
  return (
    <html lang="en" suppressHydrationWarning className="h-full antialiased">
      <head>
        <link rel="stylesheet" href="/css/satoshi.css" />
        <link rel="stylesheet" href="/css/general-sans.css" />
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
