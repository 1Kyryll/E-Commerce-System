"use client";
import * as TabsPrimitive from "@radix-ui/react-tabs";
import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cn } from "@/lib/cn";

const Root = TabsPrimitive.Root;

const List = forwardRef<
  React.ComponentRef<typeof TabsPrimitive.List>,
  ComponentPropsWithoutRef<typeof TabsPrimitive.List>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.List
    ref={ref}
    className={cn("inline-flex h-10 items-center justify-center gap-1 rounded-md bg-sub-accent-1/20 p-1", className)}
    {...props}
  />
));
List.displayName = "Tabs.List";

const Trigger = forwardRef<
  React.ComponentRef<typeof TabsPrimitive.Trigger>,
  ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Trigger
    ref={ref}
    className={cn(
      "inline-flex items-center justify-center rounded-sm px-3 py-1.5 text-sm font-medium text-secondary transition data-[state=active]:bg-primary data-[state=active]:shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary/40",
      className,
    )}
    {...props}
  />
));
Trigger.displayName = "Tabs.Trigger";

const Content = forwardRef<
  React.ComponentRef<typeof TabsPrimitive.Content>,
  ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Content ref={ref} className={cn("mt-3", className)} {...props} />
));
Content.displayName = "Tabs.Content";

export const Tabs = Object.assign(Root, { List, Trigger, Content });
