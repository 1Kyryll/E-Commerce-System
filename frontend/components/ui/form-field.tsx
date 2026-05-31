"use client";
import { createContext, useContext, useId, type ReactNode } from "react";
import { cn } from "@/lib/cn";

type Ctx = { id: string; describedById: string; error?: string };
const FieldCtx = createContext<Ctx | null>(null);

function useField() {
  const ctx = useContext(FieldCtx);
  if (!ctx) throw new Error("Field.* must be used inside <Field>");
  return ctx;
}

function Root({
  children,
  error,
  className,
}: {
  children: ReactNode;
  error?: string;
  className?: string;
}) {
  const id = useId();
  return (
    <FieldCtx.Provider value={{ id, describedById: `${id}-msg`, error }}>
      <div className={cn("flex flex-col gap-1.5", className)}>{children}</div>
    </FieldCtx.Provider>
  );
}

function FieldLabel({ children, className }: { children: ReactNode; className?: string }) {
  const { id } = useField();
  return (
    <label htmlFor={id} className={cn("text-sm font-medium text-secondary", className)}>
      {children}
    </label>
  );
}

function Control({ children }: { children: (props: { id: string; "aria-describedby": string; "aria-invalid": boolean }) => ReactNode }) {
  const { id, describedById, error } = useField();
  return <>{children({ id, "aria-describedby": describedById, "aria-invalid": !!error })}</>;
}

function Message({ children, className }: { children?: ReactNode; className?: string }) {
  const { describedById, error } = useField();
  const text = error ?? children;
  if (!text) return null;
  return (
    <p id={describedById} className={cn("text-sm", error ? "text-red-600" : "text-sub-accent-1", className)}>
      {text}
    </p>
  );
}

export const Field = Object.assign(Root, {
  Label: FieldLabel,
  Control,
  Message,
});
