import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

const inputVariants = cva(
  "flex w-full rounded-md border bg-primary px-3 py-2 text-secondary placeholder:text-sub-accent-1 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary/40 disabled:opacity-50 disabled:cursor-not-allowed",
  {
    variants: {
      size: { sm: "h-8 text-sm", md: "h-10", lg: "h-12 text-lg" },
      invalid: { true: "border-red-500", false: "border-sub-accent-1" },
    },
    defaultVariants: { size: "md", invalid: false },
  },
);

export type InputProps = Omit<ComponentPropsWithoutRef<"input">, "size"> &
  VariantProps<typeof inputVariants>;

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, size, invalid, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(inputVariants({ size, invalid }), className)}
      {...props}
    />
  ),
);
Input.displayName = "Input";
