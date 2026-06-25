package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cartv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/cart/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/services/cart/domain"
)

type fakeSvc struct {
	getFn            func(ctx context.Context, userID uuid.UUID) (domain.Cart, error)
	addFn            func(ctx context.Context, userID, productID uuid.UUID, qty int32) (domain.Cart, error)
	removeFn         func(ctx context.Context, userID, productID uuid.UUID) (domain.Cart, error)
	clearFn          func(ctx context.Context, userID uuid.UUID) error
	setcartitemqtyFn func(ctx context.Context, cartID, productID uuid.UUID, qty int32) error
}

func (f *fakeSvc) GetCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	return f.getFn(ctx, userID)
}
func (f *fakeSvc) AddItem(ctx context.Context, userID, productID uuid.UUID, qty int32) (domain.Cart, error) {
	return f.addFn(ctx, userID, productID, qty)
}
func (f *fakeSvc) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (domain.Cart, error) {
	return f.removeFn(ctx, userID, productID)
}
func (f *fakeSvc) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return f.clearFn(ctx, userID)
}
func (f *fakeSvc) SetItemQuantity(ctx context.Context, cartID, productID uuid.UUID, qty int32) error {
	return f.setcartitemqtyFn(ctx, cartID, productID, qty)
}

func TestGetCart_OK(t *testing.T) {
	uid, pid := uuid.New(), uuid.New()
	now := time.Now()
	svc := &fakeSvc{
		getFn: func(context.Context, uuid.UUID) (domain.Cart, error) {
			return domain.Cart{
				ID: uuid.New(), UserID: uid,
				Items: []domain.CartItem{{ProductID: pid, Quantity: 2, AddedAt: now}},
			}, nil
		},
	}
	h := New(svc)
	resp, err := h.GetCart(context.Background(), &cartv1.GetCartRequest{UserId: uid.String()})
	if err != nil {
		t.Fatalf("GetCart: %v", err)
	}
	if resp.Cart.UserId != uid.String() {
		t.Errorf("user_id = %s", resp.Cart.UserId)
	}
	if len(resp.Cart.Items) != 1 {
		t.Fatalf("items len = %d", len(resp.Cart.Items))
	}
	if resp.Cart.Items[0].ProductId != pid.String() || resp.Cart.Items[0].Quantity != 2 {
		t.Errorf("item = %+v", resp.Cart.Items[0])
	}
}

func TestGetCart_BadUserID_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.GetCart(context.Background(), &cartv1.GetCartRequest{UserId: "not-a-uuid"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestAddItem_OK(t *testing.T) {
	uid, pid := uuid.New(), uuid.New()
	svc := &fakeSvc{
		addFn: func(_ context.Context, u, p uuid.UUID, q int32) (domain.Cart, error) {
			if u != uid || p != pid || q != 3 {
				t.Errorf("repo args = (%s, %s, %d)", u, p, q)
			}
			return domain.Cart{ID: uuid.New(), UserID: uid}, nil
		},
	}
	h := New(svc)
	resp, err := h.AddItem(context.Background(), &cartv1.AddItemRequest{
		UserId: uid.String(), ProductId: pid.String(), Quantity: 3,
	})
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if resp.Cart.UserId != uid.String() {
		t.Errorf("user_id = %s", resp.Cart.UserId)
	}
}

func TestAddItem_InvalidQuantity(t *testing.T) {
	svc := &fakeSvc{
		addFn: func(context.Context, uuid.UUID, uuid.UUID, int32) (domain.Cart, error) {
			return domain.Cart{}, domain.ErrInvalidQuantity
		},
	}
	h := New(svc)
	_, err := h.AddItem(context.Background(), &cartv1.AddItemRequest{
		UserId: uuid.New().String(), ProductId: uuid.New().String(), Quantity: 0,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestAddItem_ProductNotFound(t *testing.T) {
	svc := &fakeSvc{
		addFn: func(context.Context, uuid.UUID, uuid.UUID, int32) (domain.Cart, error) {
			return domain.Cart{}, domain.ErrProductNotFound
		},
	}
	h := New(svc)
	_, err := h.AddItem(context.Background(), &cartv1.AddItemRequest{
		UserId: uuid.New().String(), ProductId: uuid.New().String(), Quantity: 1,
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestAddItem_BadProductID_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.AddItem(context.Background(), &cartv1.AddItemRequest{
		UserId: uuid.New().String(), ProductId: "not-a-uuid", Quantity: 1,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestAddItem_ServiceError_Internal(t *testing.T) {
	svc := &fakeSvc{
		addFn: func(context.Context, uuid.UUID, uuid.UUID, int32) (domain.Cart, error) {
			return domain.Cart{}, errors.New("db down")
		},
	}
	h := New(svc)
	_, err := h.AddItem(context.Background(), &cartv1.AddItemRequest{
		UserId: uuid.New().String(), ProductId: uuid.New().String(), Quantity: 1,
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestRemoveItem_OK(t *testing.T) {
	uid, pid := uuid.New(), uuid.New()
	svc := &fakeSvc{
		removeFn: func(_ context.Context, u, p uuid.UUID) (domain.Cart, error) {
			if u != uid || p != pid {
				t.Errorf("args = (%s, %s)", u, p)
			}
			return domain.Cart{ID: uuid.New(), UserID: uid}, nil
		},
	}
	h := New(svc)
	resp, err := h.RemoveItem(context.Background(), &cartv1.RemoveItemRequest{
		UserId: uid.String(), ProductId: pid.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Cart.UserId != uid.String() {
		t.Errorf("user_id = %s", resp.Cart.UserId)
	}
}

func TestClearCart_OK(t *testing.T) {
	uid := uuid.New()
	called := false
	svc := &fakeSvc{
		clearFn: func(_ context.Context, u uuid.UUID) error {
			called = true
			if u != uid {
				t.Errorf("user = %s", u)
			}
			return nil
		},
	}
	h := New(svc)
	_, err := h.ClearCart(context.Background(), &cartv1.ClearCartRequest{UserId: uid.String()})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("ClearCart was not called")
	}
}

func TestSetItemQuantity_OK(t *testing.T) {
	userID, productID := uuid.New(), uuid.New()
	called := false
	svc := &fakeSvc{
		setcartitemqtyFn: func(_ context.Context, uid, pid uuid.UUID, qty int32) error {
			called = true
			if userID != uid || productID != pid {
				t.Errorf("args = (%s, %s)", uid, pid)
			}
			if qty != 3 {
				t.Errorf("qty forwarded = %d, want 3", qty)
			}
			return nil
		},
	}

	h := New(svc)
	resp, err := h.SetItemQuantity(context.Background(), &cartv1.SetItemQuantityRequest{
		UserId:    userID.String(),
		ProductId: productID.String(),
		Quantity:  3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non nil resp")
	}
	if !called {
		t.Error("service: SetItemQuantity was never called")
	}
}

func TestSetItemQuantity_BadCartID_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.SetItemQuantity(context.Background(), &cartv1.SetItemQuantityRequest{
		UserId:    "not-a-uuid",
		ProductId: uuid.New().String(),
		Quantity:  1,
	})
	if code := status.Code(err); code != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestSetItemQuantity_BadProductID_InvalidArgument(t *testing.T) {
	h := New(&fakeSvc{})
	_, err := h.SetItemQuantity(context.Background(), &cartv1.SetItemQuantityRequest{
		UserId:    uuid.New().String(),
		ProductId: "not-a-uuid",
		Quantity:  1,
	})
	if code := status.Code(err); code != codes.InvalidArgument {
		t.Errorf("code = %v", status.Code(err))
	}
}

func TestSetItemQuantity_InvalidQuantity(t *testing.T) {
	for _, qty := range []int32{0, -1, -100} {
		qty := qty
		t.Run(fmt.Sprintf("qty %d", qty), func(t *testing.T) {
			svc := &fakeSvc{
				setcartitemqtyFn: func(context.Context, uuid.UUID, uuid.UUID, int32) error {
					return domain.ErrInvalidQuantity
				},
			}
			h := New(svc)
			_, err := h.SetItemQuantity(context.Background(), &cartv1.SetItemQuantityRequest{
				UserId:    uuid.New().String(),
				ProductId: uuid.New().String(),
				Quantity:  qty,
			})
			if code := status.Code(err); code != codes.InvalidArgument {
				t.Errorf("qty %d: got gRPC code %v but want InvalidArgument", qty, code)
			}
		})
	}
}

func TestSetItemQuantity_ProductNotFound(t *testing.T) {
	svc := &fakeSvc{
		setcartitemqtyFn: func(context.Context, uuid.UUID, uuid.UUID, int32) error {
			return domain.ErrProductNotFound
		},
	}
	h := New(svc)

	_, err := h.SetItemQuantity(context.Background(), &cartv1.SetItemQuantityRequest{
		UserId: uuid.New().String(), ProductId: uuid.New().String(),
		Quantity: 1,
	})
	if code := status.Code(err); code != codes.NotFound {
		t.Errorf("got grpc code %v but want NotFound", code)
	}
}

func TestSetItemQuantity_ServiceError_Internal(t *testing.T) {
	svc := &fakeSvc{
		setcartitemqtyFn: func(context.Context, uuid.UUID, uuid.UUID, int32) error {
			return errors.New("db down")
		},
	}
	h := New(svc)

	_, err := h.SetItemQuantity(context.Background(), &cartv1.SetItemQuantityRequest{
		UserId: uuid.New().String(), ProductId: uuid.New().String(),
		Quantity: 1,
	})
	if code := status.Code(err); code != codes.Internal {
		t.Errorf("code = %v", status.Code(err))
	}

	if status.Convert(err).Message() == "db down" {
		t.Error("internal err detail leak to caller")
	}
}
