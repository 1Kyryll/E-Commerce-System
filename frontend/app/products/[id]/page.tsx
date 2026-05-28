import { notFound } from "next/navigation";
import { Container } from "@/components/container";
import { getProduct } from "@/features/catalog/api";
import { ProductDetail } from "@/features/catalog/components/ProductDetail";
import { ApiError } from "@/lib/api/errors";
import type { Product } from "@/features/catalog/types";

export default async function ProductPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  let product: Product;
  try {
    product = await getProduct(id);
  } catch (err) {
    if (err instanceof ApiError && err.isNotFound) notFound();
    throw err;
  }
  return (
    <Container className="py-10 flex-1">
      <ProductDetail product={product} />
    </Container>
  );
}
