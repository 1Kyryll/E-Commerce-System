package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
)

type fakeSvc struct {
	createFn func(ctx context.Context, idem, uid, pid uuid.UUID, qty int32) (domain.Reservation, error)
}

func (f *fakeSvc) CreateReservation(ctx context.Context, idem, uid, pid uuid.UUID, qty int32) (domain.Reservation, error) {
	return f.createFn(ctx, idem, uid, pid, qty)
}

func newReq(idem, uid, pid string, qty int32) *orderv1.CreateReservationRequest {
	return &orderv1.CreateReservationRequest{IdempotencyKey: idem, UserId: uid, ProductId: pid, Quantity: qty}
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

func TestPlaceOrder_Unimplemented(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.PlaceOrder(context.Background(), &orderv1.PlaceOrderRequest{})
	if status.Code(err) != codes.Unimplemented {
		t.Errorf("PlaceOrder code = %v, want Unimplemented", status.Code(err))
	}
}

func TestGetOrder_Unimplemented(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.GetOrder(context.Background(), &orderv1.GetOrderRequest{})
	if status.Code(err) != codes.Unimplemented {
		t.Errorf("GetOrder code = %v, want Unimplemented", status.Code(err))
	}
}
