import Link from "next/link";
import { AuthCard } from "@/features/auth/components/AuthCard";
import { LoginForm } from "@/features/auth/components/LoginForm";

export default async function LoginPage({ searchParams }: { searchParams: Promise<{ next?: string }> }) {
  const { next } = await searchParams;
  return (
    <AuthCard
      title="Welcome back"
      subtitle="Sign in to continue."
      footer={
        <p className="text-sm text-sub-accent-1">
          New here? <Link href="/signup" className="text-secondary underline">Create an account</Link>
        </p>
      }
    >
      <LoginForm next={next} />
    </AuthCard>
  );
}
