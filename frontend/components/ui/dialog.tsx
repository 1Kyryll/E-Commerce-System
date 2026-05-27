"use client";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/cn";

const Root = DialogPrimitive.Root;
const Trigger = DialogPrimitive.Trigger;
const Portal = DialogPrimitive.Portal;
const Close = DialogPrimitive.Close;

const Overlay = forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Overlay>,
  ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    ref={ref}
    className={cn(
      "fixed inset-0 z-50 bg-secondary/40 backdrop-blur-sm data-[state=open]:[animation:animate-in_150ms_ease-out] data-[state=closed]:[animation:animate-out_120ms_ease-in]",
      className,
    )}
    {...props}
  />
));
Overlay.displayName = "Dialog.Overlay";

const Content = forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Content>,
  ComponentPropsWithoutRef<typeof DialogPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <Portal>
    <Overlay />
    <DialogPrimitive.Content
      ref={ref}
      className={cn(
        "fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2 rounded-xl bg-primary p-6 shadow-lg data-[state=open]:[animation:animate-in_150ms_ease-out] data-[state=closed]:[animation:animate-out_120ms_ease-in]",
        className,
      )}
      {...props}
    >
      {children}
      <DialogPrimitive.Close
        aria-label="Close"
        className="absolute right-3 top-3 rounded-md p-1 text-secondary hover:bg-sub-accent-1/20"
      >
        <X className="h-4 w-4" />
      </DialogPrimitive.Close>
    </DialogPrimitive.Content>
  </Portal>
));
Content.displayName = "Dialog.Content";

function Title({ className, ...props }: ComponentPropsWithoutRef<typeof DialogPrimitive.Title>) {
  return (
    <DialogPrimitive.Title
      className={cn("text-lg font-semibold text-secondary font-secondary", className)}
      {...props}
    />
  );
}
function Description({ className, ...props }: ComponentPropsWithoutRef<typeof DialogPrimitive.Description>) {
  return <DialogPrimitive.Description className={cn("text-sm text-sub-accent-1", className)} {...props} />;
}

export const Dialog = Object.assign(Root, {
  Trigger,
  Portal,
  Close,
  Overlay,
  Content,
  Title,
  Description,
});
