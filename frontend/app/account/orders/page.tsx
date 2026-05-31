import { Container } from "@/components/container";
import { OrdersTable } from "@/features/account/components/OrdersTable";
import { listMyOrders } from "@/features/account/api";

export default async function OrdersPage() {
  const orders = await listMyOrders();
  return (
    <Container className="py-10 flex-1 max-w-3xl">
      <h1 className="text-3xl font-secondary font-semibold text-secondary mb-8">
        Your orders
      </h1>
      <OrdersTable orders={orders} />
    </Container>
  );
}
