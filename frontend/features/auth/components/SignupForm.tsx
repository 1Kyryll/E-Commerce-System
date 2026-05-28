"use client";
import { useActionState } from "react";
import { useFormStatus } from "react-dom";
import { signupAction, type AuthFormState } from "../actions";
import { Field } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

const initial: AuthFormState = { ok: false };

export function SignupForm() {
  const [state, action] = useActionState(signupAction, initial);
  return (
    <form action={action} className="flex flex-col gap-4">
      <Field error={state.fieldErrors?.name}>
        <Field.Label>Name</Field.Label>
        <Field.Control>{(p) => <Input name="name" autoComplete="name" required {...p} />}</Field.Control>
        <Field.Message />
      </Field>
      <Field error={state.fieldErrors?.email}>
        <Field.Label>Email</Field.Label>
        <Field.Control>{(p) => <Input type="email" name="email" autoComplete="email" required {...p} />}</Field.Control>
        <Field.Message />
      </Field>
      <Field error={state.fieldErrors?.password}>
        <Field.Label>Password</Field.Label>
        <Field.Control>
          {(p) => <Input type="password" name="password" autoComplete="new-password" minLength={8} required {...p} />}
        </Field.Control>
        <Field.Message>At least 8 characters.</Field.Message>
      </Field>
      {state.formError && (
        <p className="text-sm text-red-600" role="alert">{state.formError}</p>
      )}
      <Submit>Create account</Submit>
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
