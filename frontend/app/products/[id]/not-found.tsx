import Link from "next/link";
import { Container } from "@/components/container";
import { Button } from "@/components/ui/button";

export default function NotFound() {
  return (
    <Container className="py-16 text-center flex-1">
      <h1 className="text-2xl font-secondary text-secondary mb-2">
        Product not found
      </h1>
      <p className="text-sub-accent-1 mb-6">
        It may have been removed or never existed.
      </p>
      <Button asChild intent="outline">
        <Link href="/">Back to shop</Link>
      </Button>
    </Container>
  );
}
