import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cn } from "@/lib/cn";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-md font-primary font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary/40 disabled:opacity-50 disabled:pointer-events-none",
  {
    variants: {
      intent: {
        primary: "bg-action-primary text-primary hover:opacity-90",
        secondary: "bg-action-secondary text-secondary hover:opacity-90",
        ghost: "bg-transparent text-secondary hover:bg-sub-accent-1/20",
        outline: "border border-sub-accent-1 text-secondary hover:bg-sub-accent-1/10",
        danger: "bg-red-600 text-white hover:bg-red-700",
      },
      size: {
        sm: "h-8 px-3 text-sm",
        md: "h-10 px-4",
        lg: "h-12 px-6 text-lg",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: { intent: "primary", size: "md" },
  },
);

export type ButtonProps = VariantProps<typeof buttonVariants> &
  ComponentPropsWithoutRef<"button"> & { asChild?: boolean };

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ asChild, intent, size, className, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp
        ref={ref}
        className={cn(buttonVariants({ intent, size }), className)}
        {...props}
      />
    );
  },
);
Button.displayName = "Button";

export { buttonVariants };
