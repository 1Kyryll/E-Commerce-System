package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/payment"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/repo"
)

type fakePayment struct {
	called  bool
	wantAmt decimal.Decimal
	err     error
}

func (f *fakePayment) Charge(_ context.Context, req payment.ChargeRequest) (payment.ChargeResult, error) {
	f.called = true
	f.wantAmt = req.Amount
	if f.err != nil {
		return payment.ChargeResult{}, f.err
	}
	return payment.ChargeResult{ProviderRef: "ref"}, nil
}

func placeOrderRepo() *fakeRepo {
	return &fakeRepo{
		createFn: func(_ context.Context, p repo.CreateReservationParams) (domain.Reservation, error) {
			return domain.Reservation{
				ID: p.ID, IdempotencyKey: p.IdempotencyKey,
				ProductID: p.ProductID, UserID: p.UserID,
				Quantity: p.Quantity, Status: domain.ReservationActive,
				ExpiresAt: time.Now().Add(15 * time.Minute),
			}, nil
		},
		getFn:     func(context.Context, uuid.UUID) (domain.Reservation, error) { return domain.Reservation{}, nil },
		listFn:    func(context.Context, int32) ([]domain.Reservation, error) { return nil, nil },
		releaseFn: func(context.Context, uuid.UUID) error { return nil },
	}
}

func TestPlaceOrder_HappyPath_MultiItem(t *testing.T) {
	parentIdem, uid := uuid.New(), uuid.New()
	pidA, pidB := uuid.New(), uuid.New()

	r := placeOrderRepo()
	r.getOrderByKeyFn = func(context.Context, uuid.UUID) (domain.Order, error) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	prices := map[uuid.UUID]repo.ProductPrice{
		pidA: {Amount: decimal.RequireFromString("10.00"), Currency: "USD"},
		pidB: {Amount: decimal.RequireFromString("2.50"), Currency: "USD"},
	}
	r.priceFn = func(_ context.Context, pid uuid.UUID) (repo.ProductPrice, error) {
		return prices[pid], nil
	}
	gotReservationKeys := map[uuid.UUID]uuid.UUID{} // productID -> derived key
	r.createFn = func(_ context.Context, p repo.CreateReservationParams) (domain.Reservation, error) {
		gotReservationKeys[p.ProductID] = p.IdempotencyKey
		return domain.Reservation{
			ID: p.ID, IdempotencyKey: p.IdempotencyKey,
			ProductID: p.ProductID, UserID: p.UserID,
			Quantity: p.Quantity, Status: domain.ReservationActive,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}, nil
	}
	var finalParams repo.FinalizeOrderParams
	r.finalizeFn = func(_ context.Context, p repo.FinalizeOrderParams) (domain.Order, error) {
		finalParams = p
		return domain.Order{
			ID: p.OrderID, IdempotencyKey: p.IdempotencyKey, UserID: p.UserID,
			TotalAmount: p.TotalAmount, TotalCurrency: p.TotalCurrency,
			Status: domain.OrderPaid,
		}, nil
	}
	pay := &fakePayment{}
	svc := NewService(r, pay)

	order, err := svc.PlaceOrder(context.Background(), parentIdem, uid, []CheckoutItem{
		{ProductID: pidA, Quantity: 2},
		{ProductID: pidB, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("PlaceOrder: %v", err)
	}

	wantTotal := decimal.RequireFromString("22.50") // 10*2 + 2.50
	if !order.TotalAmount.Equal(wantTotal) {
		t.Errorf("total = %s, want %s", order.TotalAmount, wantTotal)
	}
	if !pay.wantAmt.Equal(wantTotal) {
		t.Errorf("charged %s, want %s", pay.wantAmt, wantTotal)
	}
	if len(finalParams.Items) != 2 {
		t.Errorf("finalize items = %d, want 2", len(finalParams.Items))
	}
	// Reservation keys must be deterministic from parent + product.
	for _, pid := range []uuid.UUID{pidA, pidB} {
		if gotReservationKeys[pid] == uuid.Nil {
			t.Errorf("no reservation key for %s", pid)
		}
		if gotReservationKeys[pid] == parentIdem {
			t.Errorf("reservation key for %s equals parent — not derived", pid)
		}
	}
}

func TestPlaceOrder_ReservationKeysDeterministic(t *testing.T) {
	parentIdem, uid, pid := uuid.New(), uuid.New(), uuid.New()
	r := placeOrderRepo()
	r.getOrderByKeyFn = func(context.Context, uuid.UUID) (domain.Order, error) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	r.priceFn = func(context.Context, uuid.UUID) (repo.ProductPrice, error) {
		return repo.ProductPrice{Amount: decimal.RequireFromString("1.00"), Currency: "USD"}, nil
	}
	var firstKey uuid.UUID
	r.createFn = func(_ context.Context, p repo.CreateReservationParams) (domain.Reservation, error) {
		firstKey = p.IdempotencyKey
		return domain.Reservation{
			ID: p.ID, IdempotencyKey: p.IdempotencyKey, ProductID: p.ProductID,
			UserID: p.UserID, Quantity: p.Quantity, Status: domain.ReservationActive,
		}, nil
	}
	r.finalizeFn = func(context.Context, repo.FinalizeOrderParams) (domain.Order, error) {
		return domain.Order{Status: domain.OrderPaid}, nil
	}
	svc := NewService(r, &fakePayment{})
	if _, err := svc.PlaceOrder(context.Background(), parentIdem, uid, []CheckoutItem{{ProductID: pid, Quantity: 1}}); err != nil {
		t.Fatalf("first PlaceOrder: %v", err)
	}

	// Same parent + product should derive the same reservation key.
	var secondKey uuid.UUID
	r.createFn = func(_ context.Context, p repo.CreateReservationParams) (domain.Reservation, error) {
		secondKey = p.IdempotencyKey
		return domain.Reservation{
			ID: p.ID, IdempotencyKey: p.IdempotencyKey, ProductID: p.ProductID,
			UserID: p.UserID, Quantity: p.Quantity, Status: domain.ReservationActive,
		}, nil
	}
	if _, err := svc.PlaceOrder(context.Background(), parentIdem, uid, []CheckoutItem{{ProductID: pid, Quantity: 1}}); err != nil {
		t.Fatalf("second PlaceOrder: %v", err)
	}
	if firstKey != secondKey {
		t.Errorf("derived keys differ: %s vs %s", firstKey, secondKey)
	}
}

func TestPlaceOrder_IdempotencyReplay_ReturnsExistingOrder(t *testing.T) {
	parentIdem, uid, pid := uuid.New(), uuid.New(), uuid.New()
	existing := domain.Order{ID: uuid.New(), IdempotencyKey: parentIdem, UserID: uid, Status: domain.OrderPaid}
	r := placeOrderRepo()
	r.getOrderByKeyFn = func(context.Context, uuid.UUID) (domain.Order, error) { return existing, nil }
	pay := &fakePayment{}
	svc := NewService(r, pay)

	got, err := svc.PlaceOrder(context.Background(), parentIdem, uid, []CheckoutItem{{ProductID: pid, Quantity: 1}})
	if err != nil {
		t.Fatalf("PlaceOrder: %v", err)
	}
	if got.ID != existing.ID {
		t.Errorf("got %s, want %s", got.ID, existing.ID)
	}
	if pay.called {
		t.Errorf("payment must not be called on replay")
	}
}

func TestPlaceOrder_PaymentDeclined_ReleasesAllReservations(t *testing.T) {
	pidA, pidB := uuid.New(), uuid.New()
	r := placeOrderRepo()
	r.getOrderByKeyFn = func(context.Context, uuid.UUID) (domain.Order, error) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	r.priceFn = func(context.Context, uuid.UUID) (repo.ProductPrice, error) {
		return repo.ProductPrice{Amount: decimal.RequireFromString("1.00"), Currency: "USD"}, nil
	}
	createdIDs := []uuid.UUID{}
	r.createFn = func(_ context.Context, p repo.CreateReservationParams) (domain.Reservation, error) {
		createdIDs = append(createdIDs, p.ID)
		return domain.Reservation{
			ID: p.ID, IdempotencyKey: p.IdempotencyKey, ProductID: p.ProductID,
			UserID: p.UserID, Quantity: p.Quantity, Status: domain.ReservationActive,
		}, nil
	}
	releasedIDs := []uuid.UUID{}
	r.releaseFn = func(_ context.Context, id uuid.UUID) error { releasedIDs = append(releasedIDs, id); return nil }

	pay := &fakePayment{err: payment.ErrDeclined}
	svc := NewService(r, pay)

	_, err := svc.PlaceOrder(context.Background(), uuid.New(), uuid.New(), []CheckoutItem{
		{ProductID: pidA, Quantity: 1},
		{ProductID: pidB, Quantity: 1},
	})
	if !errors.Is(err, domain.ErrPaymentDeclined) {
		t.Fatalf("err = %v, want ErrPaymentDeclined", err)
	}
	if len(releasedIDs) != len(createdIDs) {
		t.Errorf("released %d, created %d — must release all", len(releasedIDs), len(createdIDs))
	}
}

func TestPlaceOrder_InsufficientInventory_ReleasesEarlierReservations(t *testing.T) {
	pidA, pidB := uuid.New(), uuid.New()
	r := placeOrderRepo()
	r.getOrderByKeyFn = func(context.Context, uuid.UUID) (domain.Order, error) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	r.priceFn = func(context.Context, uuid.UUID) (repo.ProductPrice, error) {
		return repo.ProductPrice{Amount: decimal.RequireFromString("1.00"), Currency: "USD"}, nil
	}
	createdIDs := []uuid.UUID{}
	r.createFn = func(_ context.Context, p repo.CreateReservationParams) (domain.Reservation, error) {
		if p.ProductID == pidB {
			return domain.Reservation{}, domain.ErrInsufficientInventory
		}
		createdIDs = append(createdIDs, p.ID)
		return domain.Reservation{
			ID: p.ID, IdempotencyKey: p.IdempotencyKey, ProductID: p.ProductID,
			UserID: p.UserID, Quantity: p.Quantity, Status: domain.ReservationActive,
		}, nil
	}
	releasedIDs := []uuid.UUID{}
	r.releaseFn = func(_ context.Context, id uuid.UUID) error { releasedIDs = append(releasedIDs, id); return nil }

	svc := NewService(r, &fakePayment{})
	_, err := svc.PlaceOrder(context.Background(), uuid.New(), uuid.New(), []CheckoutItem{
		{ProductID: pidA, Quantity: 1},
		{ProductID: pidB, Quantity: 1},
	})
	if !errors.Is(err, domain.ErrInsufficientInventory) {
		t.Fatalf("err = %v, want ErrInsufficientInventory", err)
	}
	if len(releasedIDs) != 1 || releasedIDs[0] != createdIDs[0] {
		t.Errorf("must release the successful reservation; released=%v created=%v", releasedIDs, createdIDs)
	}
}

func TestPlaceOrder_DuplicateProduct_InvalidArgument(t *testing.T) {
	pid := uuid.New()
	svc := NewService(placeOrderRepo(), &fakePayment{})
	_, err := svc.PlaceOrder(context.Background(), uuid.New(), uuid.New(), []CheckoutItem{
		{ProductID: pid, Quantity: 1},
		{ProductID: pid, Quantity: 2},
	})
	if !errors.Is(err, domain.ErrInvalidQuantity) && err == nil {
		t.Fatalf("err = %v, want a duplicate-product error", err)
	}
	// Accept either a dedicated sentinel or ErrInvalidQuantity-like; the
	// service must define ErrDuplicateProduct — assert on that:
	if !errors.Is(err, domain.ErrDuplicateProduct) {
		t.Errorf("err = %v, want ErrDuplicateProduct", err)
	}
}

func TestPlaceOrder_EmptyItems_InvalidArgument(t *testing.T) {
	svc := NewService(placeOrderRepo(), &fakePayment{})
	_, err := svc.PlaceOrder(context.Background(), uuid.New(), uuid.New(), nil)
	if !errors.Is(err, domain.ErrEmptyCheckout) {
		t.Errorf("err = %v, want ErrEmptyCheckout", err)
	}
}

func TestPlaceOrder_InvalidQuantity(t *testing.T) {
	svc := NewService(placeOrderRepo(), &fakePayment{})
	_, err := svc.PlaceOrder(context.Background(), uuid.New(), uuid.New(), []CheckoutItem{{ProductID: uuid.New(), Quantity: 0}})
	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v, want ErrInvalidQuantity", err)
	}
}

func TestPlaceOrder_ProductNotFound(t *testing.T) {
	r := placeOrderRepo()
	r.getOrderByKeyFn = func(context.Context, uuid.UUID) (domain.Order, error) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	r.priceFn = func(context.Context, uuid.UUID) (repo.ProductPrice, error) {
		return repo.ProductPrice{}, domain.ErrProductNotFound
	}
	svc := NewService(r, &fakePayment{})
	_, err := svc.PlaceOrder(context.Background(), uuid.New(), uuid.New(), []CheckoutItem{{ProductID: uuid.New(), Quantity: 1}})
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Errorf("err = %v, want ErrProductNotFound", err)
	}
}
