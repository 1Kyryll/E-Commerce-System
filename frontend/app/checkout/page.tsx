import { Container } from "@/components/container";
import { Checkout } from "@/features/checkout/components/Checkout";

export default function CheckoutPage() {
  return (
    <Container className="py-10 flex-1 max-w-3xl">
      <h1 className="text-3xl font-secondary font-semibold text-secondary mb-8">
        Checkout
      </h1>
      <Checkout />
    </Container>
  );
}
