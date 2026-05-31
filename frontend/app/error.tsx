"use client";
import { useEffect } from "react";
import { Container } from "@/components/container";
import { Button } from "@/components/ui/button";
import { AlertCircle } from "@/components/icons";

export default function Error({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  useEffect(() => {
    console.error(error);
  }, [error]);
  return (
    <Container className="py-16 text-center flex-1">
      <AlertCircle className="h-10 w-10 mx-auto text-red-600 mb-3" />
      <h1 className="text-2xl font-secondary text-secondary mb-2">Something went wrong</h1>
      <p className="text-sub-accent-1 mb-6">Please try again. If it keeps happening, refresh the page.</p>
      <Button onClick={() => reset()}>Try again</Button>
    </Container>
  );
}
