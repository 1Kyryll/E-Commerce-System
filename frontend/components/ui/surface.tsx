import { forwardRef, type HTMLAttributes } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

const surface = cva("bg-primary", {
  variants: {
    elevation: {
      flat: "",
      raised: "shadow-sm",
      floating: "shadow-md",
    },
    padded: { true: "p-4", false: "" },
    rounded: { true: "rounded-xl", false: "" },
  },
  defaultVariants: { elevation: "flat", padded: false, rounded: false },
});

type SurfaceProps = HTMLAttributes<HTMLDivElement> & VariantProps<typeof surface>;

export const Surface = forwardRef<HTMLDivElement, SurfaceProps>(
  ({ className, elevation, padded, rounded, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(surface({ elevation, padded, rounded }), className)}
      {...props}
    />
  ),
);
Surface.displayName = "Surface";
