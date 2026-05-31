import Link from "next/link";
import { Container } from "@/components/container";
import { Button } from "@/components/ui/button";
import { ProfileCard } from "@/features/account/components/ProfileCard";
import { getMe } from "@/features/account/api";

export default async function AccountPage() {
  const me = await getMe();
  return (
    <Container className="py-10 flex-1 max-w-3xl">
      <h1 className="text-3xl font-secondary font-semibold text-secondary mb-8">
        Account
      </h1>
      <div className="flex flex-col gap-4">
        <ProfileCard name={me.name} email={me.email} />
        <Button asChild intent="outline" className="self-start">
          <Link href="/account/orders">View past orders</Link>
        </Button>
      </div>
    </Container>
  );
}
