import type { ReactNode } from "react";
import { Card } from "@/components/ui/card";

export function AuthCard({ title, subtitle, children, footer }: {
  title: string;
  subtitle?: string;
  children: ReactNode;
  footer?: ReactNode;
}) {
  return (
    <Card className="w-full max-w-md mx-auto">
      <Card.Header>
        <h1 className="text-2xl font-secondary font-semibold text-secondary">{title}</h1>
        {subtitle && <p className="text-sm text-sub-accent-1 mt-1">{subtitle}</p>}
      </Card.Header>
      <Card.Body>{children}</Card.Body>
      {footer && <Card.Footer>{footer}</Card.Footer>}
    </Card>
  );
}
