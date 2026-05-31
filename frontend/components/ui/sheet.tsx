"use client";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { X } from "lucide-react";
import { cn } from "@/lib/cn";

const Root = DialogPrimitive.Root;
const Trigger = DialogPrimitive.Trigger;
const Close = DialogPrimitive.Close;

const sheetVariants = cva(
  "fixed z-50 bg-primary shadow-xl data-[state=open]:[animation:animate-in_180ms_ease-out] data-[state=closed]:[animation:animate-out_140ms_ease-in] flex flex-col",
  {
    variants: {
      side: {
        right: "right-0 top-0 h-full w-full max-w-md border-l border-sub-accent-1/50",
        left:  "left-0 top-0 h-full w-full max-w-md border-r border-sub-accent-1/50",
        top:    "top-0 left-0 w-full max-h-[80vh] border-b border-sub-accent-1/50",
        bottom: "bottom-0 left-0 w-full max-h-[80vh] border-t border-sub-accent-1/50",
      },
    },
    defaultVariants: { side: "right" },
  },
);

type ContentProps = ComponentPropsWithoutRef<typeof DialogPrimitive.Content> &
  VariantProps<typeof sheetVariants>;

const Content = forwardRef<React.ComponentRef<typeof DialogPrimitive.Content>, ContentProps>(
  ({ side, className, children, ...props }, ref) => (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-secondary/40 backdrop-blur-sm data-[state=open]:[animation:animate-in_180ms_ease-out] data-[state=closed]:[animation:animate-out_140ms_ease-in]" />
      <DialogPrimitive.Content ref={ref} className={cn(sheetVariants({ side }), className)} {...props}>
        {children}
        <DialogPrimitive.Close
          aria-label="Close"
          className="absolute right-3 top-3 rounded-md p-1 text-secondary hover:bg-sub-accent-1/20"
        >
          <X className="h-4 w-4" />
        </DialogPrimitive.Close>
      </DialogPrimitive.Content>
    </DialogPrimitive.Portal>
  ),
);
Content.displayName = "Sheet.Content";

function Header({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("p-4 border-b border-sub-accent-1/50", className)} {...props} />;
}
function Body({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex-1 overflow-y-auto p-4", className)} {...props} />;
}
function Footer({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("p-4 border-t border-sub-accent-1/50", className)} {...props} />;
}
function Title({ className, ...props }: ComponentPropsWithoutRef<typeof DialogPrimitive.Title>) {
  return <DialogPrimitive.Title className={cn("text-lg font-semibold font-secondary text-secondary", className)} {...props} />;
}

export const Sheet = Object.assign(Root, { Trigger, Close, Content, Header, Body, Footer, Title });
