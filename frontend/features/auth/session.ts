import "server-only";
import { cache } from "react";
import { serverApi, unwrap } from "@/lib/api/server";
import { ApiError } from "@/lib/api/errors";

export const getCurrentUser = cache(async () => {
  try {
    const api = serverApi();
    const res = await api.GET("/me");
    return await unwrap(res);
  } catch (err) {
    if (err instanceof ApiError && err.isAuthRequired) return null;
    throw err;
  }
});

export async function requireUser() {
  const user = await getCurrentUser();
  if (!user) throw new ApiError(401, "auth_required", "Sign in to continue.");
  return user;
}
