import "server-only";
import { serverApi, unwrap } from "@/lib/api/server";
import type { Product, ProductList } from "./types";

export async function listProducts(
  params: { cursor?: string; pageSize?: number } = {},
): Promise<ProductList> {
  const api = serverApi();
  const res = await api.GET("/products", {
    params: {
      query: {
        cursor: params.cursor,
        page_size: params.pageSize ?? 20,
      },
    },
    next: { tags: ["products"], revalidate: 60 },
  });
  return await unwrap(res);
}

export async function getProduct(id: string): Promise<Product> {
  const api = serverApi();
  const res = await api.GET("/products/{id}", {
    params: { path: { id } },
    next: { tags: [`product:${id}`], revalidate: 60 },
  });
  return await unwrap(res);
}
