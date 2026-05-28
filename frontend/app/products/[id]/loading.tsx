import { Container } from "@/components/container";
import { Skeleton } from "@/components/ui/skeleton";

export default function Loading() {
  return (
    <Container className="py-10">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        <Skeleton className="aspect-square" />
        <div className="flex flex-col gap-4">
          <Skeleton className="h-9 w-2/3" />
          <Skeleton className="h-6 w-1/4" />
          <Skeleton className="h-24" />
          <Skeleton className="h-10 w-40" />
        </div>
      </div>
    </Container>
  );
}
