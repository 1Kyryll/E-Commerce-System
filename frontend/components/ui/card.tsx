import { forwardRef, type HTMLAttributes, type ReactNode } from "react";
import { cn } from "@/lib/cn";

const Root = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        "rounded-xl border border-sub-accent-1/50 bg-primary shadow-sm overflow-hidden",
        className,
      )}
      {...props}
    />
  ),
);
Root.displayName = "Card";

function Header({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn("p-4 border-b border-sub-accent-1/50", className)}>{children}</div>;
}
function Body({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn("p-4", className)}>{children}</div>;
}
function Footer({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn("p-4 border-t border-sub-accent-1/50", className)}>{children}</div>;
}

export const Card = Object.assign(Root, { Header, Body, Footer });
