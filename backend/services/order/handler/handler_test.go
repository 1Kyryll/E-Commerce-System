package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/service"
)

type fakeSvc struct {
	createFn   func(ctx context.Context, idem, uid, pid uuid.UUID, qty int32) (domain.Reservation, error)
	placeFn    func(ctx context.Context, idem, uid uuid.UUID, items []service.CheckoutItem) (domain.Order, error)
	getOrderFn func(ctx context.Context, id uuid.UUID) (domain.Order, error)
}

func (f *fakeSvc) CreateReservation(ctx context.Context, idem, uid, pid uuid.UUID, qty int32) (domain.Reservation, error) {
	return f.createFn(ctx, idem, uid, pid, qty)
}
func (f *fakeSvc) PlaceOrder(ctx context.Context, idem, uid uuid.UUID, items []service.CheckoutItem) (domain.Order, error) {
	if f.placeFn == nil {
		return domain.Order{}, nil
	}
	return f.placeFn(ctx, idem, uid, items)
}
func (f *fakeSvc) GetOrder(ctx context.Context, id uuid.UUID) (domain.Order, error) {
	if f.getOrderFn == nil {
		return domain.Order{}, nil
	}
	return f.getOrderFn(ctx, id)
}

func newReq(idem, uid, pid string, qty int32) *orderv1.CreateReservationRequest {
	return &orderv1.CreateReservationRequest{IdempotencyKey: idem, UserId: uid, ProductId: pid, Quantity: qty}
}

func newPlaceReq(idem, uid string, items ...*orderv1.CheckoutItem) *orderv1.PlaceOrderRequest {
	return &orderv1.PlaceOrderRequest{IdempotencyKey: idem, UserId: uid, Items: items}
}

func TestCreateReservation_OK(t *testing.T) {
	idem, uid, pid := uuid.New(), uuid.New(), uuid.New()
	svc := &fakeSvc{
		createFn: func(_ context.Context, i, u, p uuid.UUID, q int32) (domain.Reservation, error) {
			if i != idem || u != uid || p != pid || q != 3 {
				t.Errorf("service args wrong")
			}
			return domain.Reservation{
				ID: uuid.New(), IdempotencyKey: i, ProductID: p, UserID: u,
				Quantity: q, Status: domain.ReservationActive,
				ExpiresAt: time.Now().Add(15 * time.Minute),
				CreatedAt: time.Now(),
			}, nil
		},
	}
	h := New(svc)
	resp, err := h.CreateReservation(context.Background(), newReq(idem.String(), uid.String(), pid.String(), 3))
	if err != nil {
		t.Fatalf("CreateReservation: %v", err)
	}
	if resp.Reservation.Quantity != 3 || resp.Reservation.Status != "active" {
		t.Errorf("resp = %+v", resp.Reservation)
	}
}

func TestCreateReservation_BadIdempotencyKey_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.CreateReservation(context.Background(), newReq("not-a-uuid", uuid.New().String(), uuid.New().String(), 1))
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestCreateReservation_BadUserID_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.CreateReservation(context.Background(), newReq(uuid.New().String(), "not-a-uuid", uuid.New().String(), 1))
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestCreateReservation_BadProductID_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.CreateReservation(context.Background(), newReq(uuid.New().String(), uuid.New().String(), "not-a-uuid", 1))
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestCreateReservation_InvalidQuantity(t *testing.T) {
	svc := &fakeSvc{
		createFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int32) (domain.Reservation, error) {
			return domain.Reservation{}, domain.ErrInvalidQuantity
		},
	}
	h := New(svc)
	_, err := h.CreateReservation(context.Background(), newReq(uuid.New().String(), uuid.New().String(), uuid.New().String(), 0))
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestCreateReservation_InsufficientInventory_FailedPrecondition(t *testing.T) {
	svc := &fakeSvc{
		createFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int32) (domain.Reservation, error) {
			return domain.Reservation{}, domain.ErrInsufficientInventory
		},
	}
	h := New(svc)
	_, err := h.CreateReservation(context.Background(), newReq(uuid.New().String(), uuid.New().String(), uuid.New().String(), 1))
	if status.Code(err) != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", status.Code(err))
	}
}

func TestCreateReservation_ProductNotFound_NotFound(t *testing.T) {
	svc := &fakeSvc{
		createFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int32) (domain.Reservation, error) {
			return domain.Reservation{}, domain.ErrProductNotFound
		},
	}
	h := New(svc)
	_, err := h.CreateReservation(context.Background(), newReq(uuid.New().String(), uuid.New().String(), uuid.New().String(), 1))
	if status.Code(err) != codes.NotFound {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestCreateReservation_ServiceError_Internal(t *testing.T) {
	svc := &fakeSvc{
		createFn: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int32) (domain.Reservation, error) {
			return domain.Reservation{}, errors.New("db down")
		},
	}
	h := New(svc)
	_, err := h.CreateReservation(context.Background(), newReq(uuid.New().String(), uuid.New().String(), uuid.New().String(), 1))
	if status.Code(err) != codes.Internal {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestHandler_PlaceOrder_OK_MultiItem(t *testing.T) {
	idem, uid, pidA, pidB := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	h := New(&fakeSvc{placeFn: func(_ context.Context, i, u uuid.UUID, items []service.CheckoutItem) (domain.Order, error) {
		if i != idem || u != uid {
			t.Errorf("service args wrong: i=%s u=%s", i, u)
		}
		if len(items) != 2 || items[0].ProductID != pidA || items[1].ProductID != pidB {
			t.Errorf("items wrong: %+v", items)
		}
		return domain.Order{
			ID: uuid.New(), IdempotencyKey: i, UserID: u, Status: domain.OrderPaid,
			TotalAmount:   decimal.RequireFromString("22.50"),
			TotalCurrency: "USD",
			Items: []domain.OrderItem{
				{ProductID: pidA, Quantity: 2, UnitPriceAmount: decimal.RequireFromString("10.00"), UnitPriceCurrency: "USD"},
				{ProductID: pidB, Quantity: 1, UnitPriceAmount: decimal.RequireFromString("2.50"), UnitPriceCurrency: "USD"},
			},
		}, nil
	}})
	resp, err := h.PlaceOrder(context.Background(), newPlaceReq(idem.String(), uid.String(),
		&orderv1.CheckoutItem{ProductId: pidA.String(), Quantity: 2},
		&orderv1.CheckoutItem{ProductId: pidB.String(), Quantity: 1},
	))
	if err != nil {
		t.Fatalf("PlaceOrder: %v", err)
	}
	if resp.Order.Status != "paid" || resp.Order.Total.Amount != "22.5" && resp.Order.Total.Amount != "22.50" {
		t.Errorf("resp = %+v", resp.Order)
	}
	if len(resp.Order.Items) != 2 {
		t.Errorf("items = %d, want 2", len(resp.Order.Items))
	}
}

func TestHandler_PlaceOrder_BadIdem_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.PlaceOrder(context.Background(), newPlaceReq("nope", uuid.New().String(),
		&orderv1.CheckoutItem{ProductId: uuid.New().String(), Quantity: 1}))
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestHandler_PlaceOrder_BadItemProductID_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.PlaceOrder(context.Background(), newPlaceReq(uuid.New().String(), uuid.New().String(),
		&orderv1.CheckoutItem{ProductId: "nope", Quantity: 1}))
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestHandler_PlaceOrder_PaymentDeclined_FailedPrecondition(t *testing.T) {
	h := New(&fakeSvc{placeFn: func(context.Context, uuid.UUID, uuid.UUID, []service.CheckoutItem) (domain.Order, error) {
		return domain.Order{}, domain.ErrPaymentDeclined
	}})
	_, err := h.PlaceOrder(context.Background(), newPlaceReq(uuid.New().String(), uuid.New().String(),
		&orderv1.CheckoutItem{ProductId: uuid.New().String(), Quantity: 1}))
	if status.Code(err) != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", status.Code(err))
	}
}

func TestHandler_PlaceOrder_InsufficientInventory_FailedPrecondition(t *testing.T) {
	h := New(&fakeSvc{placeFn: func(context.Context, uuid.UUID, uuid.UUID, []service.CheckoutItem) (domain.Order, error) {
		return domain.Order{}, domain.ErrInsufficientInventory
	}})
	_, err := h.PlaceOrder(context.Background(), newPlaceReq(uuid.New().String(), uuid.New().String(),
		&orderv1.CheckoutItem{ProductId: uuid.New().String(), Quantity: 1}))
	if status.Code(err) != codes.FailedPrecondition {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestHandler_PlaceOrder_ProductNotFound_NotFound(t *testing.T) {
	h := New(&fakeSvc{placeFn: func(context.Context, uuid.UUID, uuid.UUID, []service.CheckoutItem) (domain.Order, error) {
		return domain.Order{}, domain.ErrProductNotFound
	}})
	_, err := h.PlaceOrder(context.Background(), newPlaceReq(uuid.New().String(), uuid.New().String(),
		&orderv1.CheckoutItem{ProductId: uuid.New().String(), Quantity: 1}))
	if status.Code(err) != codes.NotFound {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestHandler_PlaceOrder_DuplicateProduct_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{placeFn: func(context.Context, uuid.UUID, uuid.UUID, []service.CheckoutItem) (domain.Order, error) {
		return domain.Order{}, domain.ErrDuplicateProduct
	}})
	_, err := h.PlaceOrder(context.Background(), newPlaceReq(uuid.New().String(), uuid.New().String(),
		&orderv1.CheckoutItem{ProductId: uuid.New().String(), Quantity: 1}))
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestHandler_GetOrder_OK(t *testing.T) {
	id := uuid.New()
	h := New(&fakeSvc{getOrderFn: func(_ context.Context, got uuid.UUID) (domain.Order, error) {
		return domain.Order{ID: got, Status: domain.OrderPaid, TotalAmount: decimal.Zero, TotalCurrency: "USD"}, nil
	}})
	resp, err := h.GetOrder(context.Background(), &orderv1.GetOrderRequest{Id: id.String()})
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if resp.Order.Id != id.String() {
		t.Errorf("got %s", resp.Order.Id)
	}
}

func TestHandler_GetOrder_BadID_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.GetOrder(context.Background(), &orderv1.GetOrderRequest{Id: "nope"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestHandler_GetOrder_NotFound(t *testing.T) {
	h := New(&fakeSvc{getOrderFn: func(context.Context, uuid.UUID) (domain.Order, error) {
		return domain.Order{}, domain.ErrOrderNotFound
	}})
	_, err := h.GetOrder(context.Background(), &orderv1.GetOrderRequest{Id: uuid.New().String()})
	if status.Code(err) != codes.NotFound {
		t.Errorf("code = %v", status.Code(err))
	}
}
