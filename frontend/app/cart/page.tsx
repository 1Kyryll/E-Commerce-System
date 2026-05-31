import Link from "next/link";
import { Container } from "@/components/container";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { formatMoney } from "@/lib/money";
import { getCart, type Cart, type CartItem } from "@/features/cart/api";
import { getProduct } from "@/features/catalog/api";
import type { Product } from "@/features/catalog/types";
import type { components } from "@/lib/types";

type Money = components["schemas"]["Money"];

async function loadCart(): Promise<Cart> {
  try {
    return await getCart();
  } catch {
    return {
      id: "00000000-0000-0000-0000-000000000000",
      user_id: "00000000-0000-0000-0000-000000000000",
      items: [],
    };
  }
}

async function loadLine(it: CartItem): Promise<{
  item: CartItem;
  product: Product | null;
}> {
  try {
    const product = await getProduct(it.product_id);
    return { item: it, product };
  } catch {
    return { item: it, product: null };
  }
}

export default async function CartPage() {
  const cart = await loadCart();
  const items = cart.items ?? [];
  if (items.length === 0) {
    return (
      <Container className="py-16 text-center flex-1">
        <h1 className="text-2xl font-secondary text-secondary mb-2">
          Your cart is empty
        </h1>
        <p className="text-sub-accent-1 mb-6">
          Find something you like in the shop.
        </p>
        <Button asChild>
          <Link href="/">Browse products</Link>
        </Button>
      </Container>
    );
  }

  const lines = await Promise.all(items.map(loadLine));
  const currency = lines.find((l) => l.product)?.product?.price.currency ?? "USD";
  const subtotal = lines.reduce((acc, l) => {
    if (!l.product) return acc;
    return acc + Number(l.product.price.amount) * (l.item.quantity ?? 0);
  }, 0);
  const subtotalMoney: Money = { amount: subtotal.toFixed(4), currency };

  return (
    <Container className="py-10 flex-1">
      <h1 className="text-3xl font-secondary font-semibold text-secondary mb-8">
        Your cart
      </h1>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <ul className="lg:col-span-2 flex flex-col gap-3">
          {lines.map(({ item, product }) => (
            <Card key={item.product_id}>
              <Card.Body>
                <div className="flex items-center gap-4">
                  <div className="flex-1">
                    <p className="font-medium text-secondary">
                      {product?.name ?? item.product_id}
                    </p>
                    {product && (
                      <p className="text-sub-accent-1 text-sm">
                        {formatMoney(product.price)}
                      </p>
                    )}
                  </div>
                  <span className="text-secondary">× {item.quantity ?? 0}</span>
                </div>
              </Card.Body>
            </Card>
          ))}
        </ul>
        <Card>
          <Card.Header>
            <h2 className="font-secondary text-lg">Summary</h2>
          </Card.Header>
          <Card.Body className="flex items-center justify-between">
            <span>Subtotal</span>
            <span className="font-secondary">{formatMoney(subtotalMoney)}</span>
          </Card.Body>
          <Card.Footer>
            <Button asChild className="w-full">
              <Link href="/checkout">Checkout</Link>
            </Button>
          </Card.Footer>
        </Card>
      </div>
    </Container>
  );
}
