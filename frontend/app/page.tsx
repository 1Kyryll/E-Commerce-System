import { Container } from "@/components/container";
import { listProducts } from "@/features/catalog/api";
import { ProductGrid } from "@/features/catalog/components/ProductGrid";
import { Pagination } from "@/features/catalog/components/Pagination";

export default async function Home({
  searchParams,
}: {
  searchParams: Promise<{ cursor?: string }>;
}) {
  const { cursor } = await searchParams;
  const list = await listProducts({ cursor });
  return (
    <Container className="py-10 flex-1">
      <h1 className="text-3xl font-secondary font-semibold text-secondary mb-8">
        Shop
      </h1>
      <ProductGrid items={list.products ?? []} />
      <Pagination nextCursor={list.next_page_cursor || null} />
    </Container>
  );
}
