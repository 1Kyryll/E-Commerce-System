package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/payment"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/repo"
)

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
//
//  1. Validate the items list (non-empty, positive quantities, no dup products).
//  2. Idempotency check: if an order already exists for the parent key,
//     return it (true replay — payment + finalize already happened).
//  3. For each item: look up product price; create the per-item reservation
//     using a derived idempotency_key (UUID v5 of parent + product). The
//     existing CreateReservation is itself idempotent on that key, so a
//     mid-flight retry produces the same reservations.
//  4. Charge once for the grand total.
//  5. On payment success: FinalizeOrder transaction (consume all reservations,
//     insert order + N order_items, write outbox event).
//  6. On payment failure or mid-flight reservation failure: Release all
//     successfully-created reservations (inventory restored).
//
// Timeout / reconcile is deferred — see docs/system-design.md.
func (s *Service) PlaceOrder(ctx context.Context, idempotencyKey, userID uuid.UUID, items []CheckoutItem) (domain.Order, error) {
	if len(items) == 0 {
		return domain.Order{}, domain.ErrEmptyCheckout
	}
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, it := range items {
		if it.Quantity < 1 {
			return domain.Order{}, domain.ErrInvalidQuantity
		}
		if _, dup := seen[it.ProductID]; dup {
			return domain.Order{}, domain.ErrDuplicateProduct
		}
		seen[it.ProductID] = struct{}{}
	}

	if existing, err := s.repo.GetOrderByIdempotencyKey(ctx, idempotencyKey); err == nil {
		return existing, nil
	} else if !errors.Is(err, domain.ErrOrderNotFound) {
		return domain.Order{}, fmt.Errorf("idempotency lookup: %w", err)
	}

	// Phase 1: per-item price + reservation. Track successes for compensation.
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
				// Best-effort; cleanup worker will catch leftovers.
				_ = releaseErr
			}
		}
	}

	for _, it := range items {
		price, err := s.repo.GetProductPriceForOrder(ctx, it.ProductID)
		if err != nil {
			releaseAll()
			return domain.Order{}, err
		}
		if currency == "" {
			currency = price.Currency
		} else if currency != price.Currency {
			releaseAll()
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
			releaseAll()
			return domain.Order{}, err
		}
		if reservation.Status != domain.ReservationActive {
			releaseAll()
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

	if _, err := s.payment.Charge(ctx, payment.ChargeRequest{
		IdempotencyKey: idempotencyKey,
		Amount:         total,
		Currency:       currency,
	}); err != nil {
		releaseAll()
		if errors.Is(err, payment.ErrDeclined) || errors.Is(err, payment.ErrTimeout) {
			return domain.Order{}, domain.ErrPaymentDeclined
		}
		return domain.Order{}, fmt.Errorf("charge: %w", err)
	}

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

	finalized, err := s.repo.FinalizeOrder(ctx, repo.FinalizeOrderParams{
		OrderID:        uuid.New(),
		IdempotencyKey: idempotencyKey,
		UserID:         userID,
		TotalAmount:    total,
		TotalCurrency:  currency,
		Items:          finalizeItems,
		OutboxEventID:  uuid.New(),
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("finalize: %w", err)
	}
	return finalized, nil
}
