package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cartv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/cart/v1"
	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

type fakeOrderClient struct {
	placeFn func(ctx context.Context, in *orderv1.PlaceOrderRequest, opts ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error)
	getFn   func(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
}

func (f *fakeOrderClient) PlaceOrder(ctx context.Context, in *orderv1.PlaceOrderRequest, opts ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error) {
	return f.placeFn(ctx, in, opts...)
}
func (f *fakeOrderClient) GetOrder(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
	return f.getFn(ctx, in, opts...)
}

func authed(uid uuid.UUID, r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), middleware.UserIDContextKey(), uid))
}

// cartWithItems builds a fakeCartClient pre-populated with the given items.
// ClearCart is recorded as a no-op success.
func cartWithItems(uid uuid.UUID, clearCalled *bool, items ...*cartv1.CartItem) *fakeCartClient {
	return &fakeCartClient{
		getFn: func(_ context.Context, in *cartv1.GetCartRequest, _ ...grpc.CallOption) (*cartv1.GetCartResponse, error) {
			return &cartv1.GetCartResponse{Cart: &cartv1.Cart{
				Id: uuid.New().String(), UserId: in.GetUserId(), Items: items,
			}}, nil
		},
		clearFn: func(context.Context, *cartv1.ClearCartRequest, ...grpc.CallOption) (*cartv1.ClearCartResponse, error) {
			if clearCalled != nil {
				*clearCalled = true
			}
			return &cartv1.ClearCartResponse{}, nil
		},
	}
}

func TestCheckout_OK_PullsItemsFromCart(t *testing.T) {
	uid, pidA, pidB, idem := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	oid := uuid.New()
	cleared := false
	cart := cartWithItems(uid, &cleared,
		&cartv1.CartItem{ProductId: pidA.String(), Quantity: 2},
		&cartv1.CartItem{ProductId: pidB.String(), Quantity: 1},
	)
	order := &fakeOrderClient{
		placeFn: func(_ context.Context, in *orderv1.PlaceOrderRequest, _ ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error) {
			if in.IdempotencyKey != idem.String() || in.UserId != uid.String() {
				t.Errorf("idem/uid wrong: %+v", in)
			}
			if len(in.Items) != 2 || in.Items[0].ProductId != pidA.String() || in.Items[1].ProductId != pidB.String() {
				t.Errorf("items not sourced from cart: %+v", in.Items)
			}
			return &orderv1.PlaceOrderResponse{Order: &orderv1.Order{
				Id: oid.String(), Status: "paid",
				Total: &orderv1.Money{Amount: "22.50", Currency: "USD"},
			}}, nil
		},
	}

	h := NewCheckoutHandlers(order, cart)
	r := httptest.NewRequest("POST", "/checkout", nil)
	r.Header.Set("Idempotency-Key", idem.String())
	w := httptest.NewRecorder()
	h.Checkout(w, authed(uid, r))

	if w.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !cleared {
		t.Errorf("cart not cleared after successful checkout")
	}
}

func TestCheckout_MissingIdempotencyKey_BadRequest(t *testing.T) {
	h := NewCheckoutHandlers(&fakeOrderClient{}, cartWithItems(uuid.New(), nil,
		&cartv1.CartItem{ProductId: uuid.New().String(), Quantity: 1}))
	r := httptest.NewRequest("POST", "/checkout", nil)
	w := httptest.NewRecorder()
	h.Checkout(w, authed(uuid.New(), r))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d", w.Code)
	}
}

func TestCheckout_EmptyCart_BadRequest(t *testing.T) {
	uid := uuid.New()
	cart := cartWithItems(uid, nil) // no items
	h := NewCheckoutHandlers(&fakeOrderClient{}, cart)
	r := httptest.NewRequest("POST", "/checkout", nil)
	r.Header.Set("Idempotency-Key", uuid.New().String())
	w := httptest.NewRecorder()
	h.Checkout(w, authed(uid, r))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d body %s", w.Code, w.Body.String())
	}
}

func TestCheckout_Unauthorized_WhenNoUserCtx(t *testing.T) {
	h := NewCheckoutHandlers(&fakeOrderClient{}, &fakeCartClient{})
	r := httptest.NewRequest("POST", "/checkout", nil)
	r.Header.Set("Idempotency-Key", uuid.New().String())
	w := httptest.NewRecorder()
	h.Checkout(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d", w.Code)
	}
}

func TestCheckout_InsufficientInventory_409_AndCartNotCleared(t *testing.T) {
	uid := uuid.New()
	cleared := false
	cart := cartWithItems(uid, &cleared,
		&cartv1.CartItem{ProductId: uuid.New().String(), Quantity: 1})
	order := &fakeOrderClient{
		placeFn: func(context.Context, *orderv1.PlaceOrderRequest, ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error) {
			return nil, status.Error(codes.FailedPrecondition, "insufficient inventory")
		},
	}
	h := NewCheckoutHandlers(order, cart)
	r := httptest.NewRequest("POST", "/checkout", nil)
	r.Header.Set("Idempotency-Key", uuid.New().String())
	w := httptest.NewRecorder()
	h.Checkout(w, authed(uid, r))
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d", w.Code)
	}
	if cleared {
		t.Errorf("cart must not be cleared when checkout fails")
	}
}
