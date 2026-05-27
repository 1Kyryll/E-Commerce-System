import { type HTMLAttributes } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

const badge = cva(
  "inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium",
  {
    variants: {
      tone: {
        neutral: "bg-sub-accent-1/30 text-secondary",
        accent: "bg-action-secondary text-secondary",
        success: "bg-action-primary text-primary",
        danger: "bg-red-100 text-red-800",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

export type BadgeProps = HTMLAttributes<HTMLSpanElement> & VariantProps<typeof badge>;

export function Badge({ tone, className, ...props }: BadgeProps) {
  return <span className={cn(badge({ tone }), className)} {...props} />;
}
