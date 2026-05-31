import Link from "next/link";
import { Container } from "@/components/container";
import { Button } from "@/components/ui/button";
import { CartNavSlot } from "@/features/cart/components/CartNavSlot";
import { getCurrentUser } from "@/features/auth/session";
import { LogOut, User } from "@/components/icons";
import { logoutAction } from "@/features/auth/actions";

export async function Nav() {
  const user = await getCurrentUser();
  return (
    <header className="border-b border-sub-accent-1/40 bg-primary">
      <Container className="flex items-center justify-between h-16">
        <Link
          href="/"
          className="font-secondary text-xl font-semibold text-secondary"
        >
          Storefront
        </Link>
        <div className="flex items-center gap-2">
          <CartNavSlot />
          {user ? (
            <>
              <Button asChild intent="ghost" size="icon">
                <Link href="/account" aria-label="Account">
                  <User className="h-5 w-5" />
                </Link>
              </Button>
              <form action={logoutAction}>
                <Button
                  type="submit"
                  intent="ghost"
                  size="icon"
                  aria-label="Sign out"
                >
                  <LogOut className="h-5 w-5" />
                </Button>
              </form>
            </>
          ) : (
            <Button asChild intent="ghost">
              <Link href="/login">Sign in</Link>
            </Button>
          )}
        </div>
      </Container>
    </header>
  );
}
