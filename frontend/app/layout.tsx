import type { Metadata } from "next";
import { ThemeProvider } from "next-themes";
import { Toaster } from "sonner";
import "./index.css";

export const metadata: Metadata = {
  title: "Storefront",
  description: "An e-commerce demo",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning className="h-full antialiased">
      <head>
        <link rel="stylesheet" href="/css/satoshi.css" />
        <link rel="stylesheet" href="/css/general-sans.css" />
      </head>
      <body className="min-h-full flex flex-col bg-primary text-secondary font-primary">
        <ThemeProvider attribute="class" defaultTheme="light" enableSystem={false}>
          {children}
          <Toaster richColors closeButton position="bottom-right" />
        </ThemeProvider>
      </body>
    </html>
  );
}
