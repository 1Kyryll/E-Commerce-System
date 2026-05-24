package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/payment"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/repo"
)

var (
	tracer = otel.Tracer("order/service")
	meter  = otel.Meter("order/service")

	placeOrderTotal metric.Int64Counter
	paymentOutcomes metric.Int64Counter
)

func init() {
	var err error
	placeOrderTotal, err = meter.Int64Counter("order.place_order.total",
		metric.WithDescription("PlaceOrder requests by outcome"))
	if err != nil {
		panic(err)
	}
	paymentOutcomes, err = meter.Int64Counter("order.payment.outcomes",
		metric.WithDescription("Payment outcomes by result"))
	if err != nil {
		panic(err)
	}
}

func recordOutcome(ctx context.Context, outcome string) {
	placeOrderTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
}

// CheckoutItem is one line in a PlaceOrder request. The service derives
// price + reservation key per item.
type CheckoutItem struct {
	ProductID uuid.UUID
	Quantity  int32
}

// deriveReservationKey computes a deterministic UUID v5 per (parent, product).
// Used as the reservation idempotency_key so retries with the same parent
// key produce the same reservation rows.
func deriveReservationKey(parent, productID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(parent.String()+":"+productID.String()))
}

// PlaceOrder runs the full multi-item purchase flow from docs/system-design.md.
func (s *Service) PlaceOrder(ctx context.Context, idempotencyKey, userID uuid.UUID, items []CheckoutItem) (domain.Order, error) {
	if len(items) == 0 {
		recordOutcome(ctx, "empty")
		return domain.Order{}, domain.ErrEmptyCheckout
	}
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, it := range items {
		if it.Quantity < 1 {
			recordOutcome(ctx, "invalid_quantity")
			return domain.Order{}, domain.ErrInvalidQuantity
		}
		if _, dup := seen[it.ProductID]; dup {
			recordOutcome(ctx, "duplicate_product")
			return domain.Order{}, domain.ErrDuplicateProduct
		}
		seen[it.ProductID] = struct{}{}
	}

	if existing, err := s.repo.GetOrderByIdempotencyKey(ctx, idempotencyKey); err == nil {
		recordOutcome(ctx, "replay")
		return existing, nil
	} else if !errors.Is(err, domain.ErrOrderNotFound) {
		recordOutcome(ctx, "error")
		return domain.Order{}, fmt.Errorf("idempotency lookup: %w", err)
	}

	type lineState struct {
		productID   uuid.UUID
		quantity    int32
		price       repo.ProductPrice
		reservation domain.Reservation
	}
	lines := make([]lineState, 0, len(items))
	currency := ""
	total := decimal.Zero

	releaseAll := func() {
		for _, ln := range lines {
			if releaseErr := s.repo.Release(ctx, ln.reservation.ID); releaseErr != nil {
				_ = releaseErr
			}
		}
	}

	ctx, reserveSpan := tracer.Start(ctx, "PlaceOrder.reserve",
		trace.WithAttributes(attribute.Int("items.count", len(items))))
	for _, it := range items {
		price, err := s.repo.GetProductPriceForOrder(ctx, it.ProductID)
		if err != nil {
			reserveSpan.SetStatus(codes.Error, err.Error())
			reserveSpan.End()
			releaseAll()
			if errors.Is(err, domain.ErrProductNotFound) {
				recordOutcome(ctx, "product_not_found")
			} else {
				recordOutcome(ctx, "error")
			}
			return domain.Order{}, err
		}
		if currency == "" {
			currency = price.Currency
		} else if currency != price.Currency {
			reserveSpan.SetStatus(codes.Error, "mixed currencies")
			reserveSpan.End()
			releaseAll()
			recordOutcome(ctx, "mixed_currency")
			return domain.Order{}, fmt.Errorf("mixed currencies in checkout: %s vs %s", currency, price.Currency)
		}

		reservation, err := s.repo.CreateReservation(ctx, repo.CreateReservationParams{
			ID:             uuid.New(),
			IdempotencyKey: deriveReservationKey(idempotencyKey, it.ProductID),
			UserID:         userID,
			ProductID:      it.ProductID,
			Quantity:       it.Quantity,
		})
		if err != nil {
			reserveSpan.SetStatus(codes.Error, err.Error())
			reserveSpan.End()
			releaseAll()
			switch {
			case errors.Is(err, domain.ErrInsufficientInventory):
				recordOutcome(ctx, "insufficient_inventory")
			case errors.Is(err, domain.ErrProductNotFound):
				recordOutcome(ctx, "product_not_found")
			default:
				recordOutcome(ctx, "error")
			}
			return domain.Order{}, err
		}
		if reservation.Status != domain.ReservationActive {
			reserveSpan.SetStatus(codes.Error, "reservation not active")
			reserveSpan.End()
			releaseAll()
			recordOutcome(ctx, "reservation_not_active")
			return domain.Order{}, domain.ErrReservationNotActive
		}

		lines = append(lines, lineState{
			productID:   it.ProductID,
			quantity:    it.Quantity,
			price:       price,
			reservation: reservation,
		})
		total = total.Add(price.Amount.Mul(decimal.NewFromInt(int64(it.Quantity))))
	}
	reserveSpan.End()

	chargeCtx, chargeSpan := tracer.Start(ctx, "PlaceOrder.charge",
		trace.WithAttributes(
			attribute.String("currency", currency),
			attribute.String("amount", total.String()),
		))
	_, chargeErr := s.payment.Charge(chargeCtx, payment.ChargeRequest{
		IdempotencyKey: idempotencyKey,
		Amount:         total,
		Currency:       currency,
	})
	paymentResult := "ok"
	if errors.Is(chargeErr, payment.ErrDeclined) {
		paymentResult = "declined"
	} else if chargeErr != nil {
		paymentResult = "error"
	}
	paymentOutcomes.Add(chargeCtx, 1, metric.WithAttributes(attribute.String("result", paymentResult)))
	if chargeErr != nil {
		chargeSpan.SetStatus(codes.Error, chargeErr.Error())
		chargeSpan.End()
		releaseAll()
		if errors.Is(chargeErr, payment.ErrDeclined) || errors.Is(chargeErr, payment.ErrTimeout) {
			recordOutcome(ctx, "payment_declined")
			return domain.Order{}, domain.ErrPaymentDeclined
		}
		recordOutcome(ctx, "error")
		return domain.Order{}, fmt.Errorf("charge: %w", chargeErr)
	}
	chargeSpan.End()

	finalizeItems := make([]repo.FinalizeOrderItem, 0, len(lines))
	for _, ln := range lines {
		finalizeItems = append(finalizeItems, repo.FinalizeOrderItem{
			OrderItemID:       uuid.New(),
			ReservationID:     ln.reservation.ID,
			ProductID:         ln.productID,
			Quantity:          ln.quantity,
			UnitPriceAmount:   ln.price.Amount,
			UnitPriceCurrency: ln.price.Currency,
		})
	}

	finalizeCtx, finalizeSpan := tracer.Start(ctx, "PlaceOrder.finalize",
		trace.WithAttributes(attribute.Int("items.count", len(finalizeItems))))
	finalized, err := s.repo.FinalizeOrder(finalizeCtx, repo.FinalizeOrderParams{
		OrderID:        uuid.New(),
		IdempotencyKey: idempotencyKey,
		UserID:         userID,
		TotalAmount:    total,
		TotalCurrency:  currency,
		Items:          finalizeItems,
		OutboxEventID:  uuid.New(),
	})
	if err != nil {
		finalizeSpan.SetStatus(codes.Error, err.Error())
		finalizeSpan.End()
		recordOutcome(ctx, "error")
		return domain.Order{}, fmt.Errorf("finalize: %w", err)
	}
	finalizeSpan.End()
	recordOutcome(ctx, "ok")
	return finalized, nil
}
