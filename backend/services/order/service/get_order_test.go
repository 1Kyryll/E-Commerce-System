package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/payment"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
)

func TestGetOrder_HappyPath(t *testing.T) {
	id := uuid.New()
	r := placeOrderRepo()
	r.getOrderByIDFn = func(_ context.Context, got uuid.UUID) (domain.Order, error) {
		if got != id {
			t.Errorf("id = %s, want %s", got, id)
		}
		return domain.Order{ID: id, Status: domain.OrderPaid}, nil
	}
	svc := NewService(r, payment.NewFakeClient())
	o, err := svc.GetOrder(context.Background(), id)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if o.ID != id {
		t.Errorf("got %s", o.ID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	r := placeOrderRepo()
	r.getOrderByIDFn = func(context.Context, uuid.UUID) (domain.Order, error) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	svc := NewService(r, payment.NewFakeClient())
	_, err := svc.GetOrder(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrOrderNotFound) {
		t.Errorf("err = %v, want ErrOrderNotFound", err)
	}
}
