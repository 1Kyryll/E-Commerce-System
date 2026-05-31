import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ArrowRight } from "@/components/icons";

export function Pagination({ nextCursor }: { nextCursor?: string | null }) {
  if (!nextCursor) return null;
  return (
    <div className="flex justify-center pt-8">
      <Button asChild intent="outline">
        <Link href={`/?cursor=${encodeURIComponent(nextCursor)}`}>
          Next page <ArrowRight className="h-4 w-4" />
        </Link>
      </Button>
    </div>
  );
}
