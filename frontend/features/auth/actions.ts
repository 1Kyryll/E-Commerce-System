"use server";
import { redirect } from "next/navigation";
import { revalidatePath } from "next/cache";
import { serverApi, unwrap, forwardSetCookies } from "@/lib/api/server";
import { ApiError, userMessageFor } from "@/lib/api/errors";
import { loginSchema, signupSchema } from "./schemas";

export type AuthFormState = {
  ok: boolean;
  formError?: string;
  fieldErrors?: Record<string, string>;
};

export async function loginAction(_prev: AuthFormState, formData: FormData): Promise<AuthFormState> {
  const parsed = loginSchema.safeParse({
    email: formData.get("email"),
    password: formData.get("password"),
  });
  if (!parsed.success) {
    return { ok: false, fieldErrors: zodErrors(parsed.error) };
  }
  try {
    const api = serverApi();
    const res = await api.POST("/auth/login", { body: parsed.data });
    await unwrap(res);
    await forwardSetCookies(res.response);
  } catch (err) {
    return { ok: false, formError: userMessageFor(err) };
  }
  const next = (formData.get("next") as string) || "/";
  revalidatePath("/", "layout");
  redirect(next);
}

export async function signupAction(_prev: AuthFormState, formData: FormData): Promise<AuthFormState> {
  const parsed = signupSchema.safeParse({
    email: formData.get("email"),
    password: formData.get("password"),
    name: formData.get("name"),
  });
  if (!parsed.success) {
    return { ok: false, fieldErrors: zodErrors(parsed.error) };
  }
  try {
    const api = serverApi();
    const res = await api.POST("/auth/signup", { body: parsed.data });
    await unwrap(res);
    await forwardSetCookies(res.response);
  } catch (err) {
    if (err instanceof ApiError && err.isConflict) {
      return { ok: false, fieldErrors: { email: "An account with that email already exists." } };
    }
    return { ok: false, formError: userMessageFor(err) };
  }
  revalidatePath("/", "layout");
  redirect("/");
}

export async function logoutAction(): Promise<void> {
  try {
    const api = serverApi();
    const res = await api.POST("/auth/logout");
    // Forward Set-Cookie so the cleared cookie (Max-Age=0) reaches the browser.
    await forwardSetCookies(res.response);
  } catch {
    // best-effort
  }
  revalidatePath("/", "layout");
  redirect("/");
}

function zodErrors(err: import("zod").ZodError): Record<string, string> {
  const out: Record<string, string> = {};
  for (const issue of err.issues) {
    const key = issue.path.join(".") || "_";
    if (!out[key]) out[key] = issue.message;
  }
  return out;
}
