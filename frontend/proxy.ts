import { NextResponse, type NextRequest } from "next/server";

const PROTECTED = ["/checkout", "/account", "/orders"];
const GATEWAY = process.env.GATEWAY_URL ?? "http://localhost:8080";

export async function proxy(req: NextRequest) {
  const { pathname } = req.nextUrl;
  if (!PROTECTED.some((p) => pathname === p || pathname.startsWith(p + "/"))) {
    return NextResponse.next();
  }

  const cookie = req.headers.get("cookie") ?? "";
  if (!cookie) return redirectToLogin(req, pathname);

  try {
    const res = await fetch(`${GATEWAY}/me`, {
      headers: { cookie },
      cache: "no-store",
    });
    if (res.ok) return NextResponse.next();
  } catch {
    // If the request to the gateway fails, we assume the user is not authenticated.
  }

  return redirectToLogin(req, pathname);
}

function redirectToLogin(req: NextRequest, pathname: string) {
  const url = req.nextUrl.clone();
  url.pathname = "/login";
  url.searchParams.set("next", pathname);
  return NextResponse.redirect(url);
}

export const config = {
  matcher: ["/checkout/:path*", "/account/:path*", "/orders/:path*"],
};
