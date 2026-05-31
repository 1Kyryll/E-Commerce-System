import type { ReactNode } from "react";
import { Container } from "@/components/container";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <Container className="flex-1 flex items-center justify-center py-16">
      {children}
    </Container>
  );
}
