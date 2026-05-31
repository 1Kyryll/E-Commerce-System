import Link from "next/link";
import { AuthCard } from "@/features/auth/components/AuthCard";
import { SignupForm } from "@/features/auth/components/SignupForm";

export default function SignupPage() {
  return (
    <AuthCard
      title="Create your account"
      footer={
        <p className="text-sm text-sub-accent-1">
          Already have one? <Link href="/login" className="text-secondary underline">Sign in</Link>
        </p>
      }
    >
      <SignupForm />
    </AuthCard>
  );
}
