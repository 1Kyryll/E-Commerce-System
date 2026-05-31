"use client";
import { useActionState } from "react";
import { useFormStatus } from "react-dom";
import { loginAction, type AuthFormState } from "../actions";
import { Field } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

const initial: AuthFormState = { ok: false };

export function LoginForm({ next }: { next?: string }) {
  const [state, action] = useActionState(loginAction, initial);
  return (
    <form action={action} className="flex flex-col gap-4">
      {next && <input type="hidden" name="next" value={next} />}
      <Field error={state.fieldErrors?.email}>
        <Field.Label>Email</Field.Label>
        <Field.Control>
          {(p) => <Input type="email" name="email" autoComplete="email" required {...p} />}
        </Field.Control>
        <Field.Message />
      </Field>
      <Field error={state.fieldErrors?.password}>
        <Field.Label>Password</Field.Label>
        <Field.Control>
          {(p) => <Input type="password" name="password" autoComplete="current-password" required {...p} />}
        </Field.Control>
        <Field.Message />
      </Field>
      {state.formError && (
        <p className="text-sm text-red-600" role="alert">{state.formError}</p>
      )}
      <Submit>Sign in</Submit>
    </form>
  );
}

function Submit({ children }: { children: React.ReactNode }) {
  const { pending } = useFormStatus();
  return (
    <Button type="submit" disabled={pending} className="w-full">
      {pending ? <Spinner /> : null}
      {children}
    </Button>
  );
}
