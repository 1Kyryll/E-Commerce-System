package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

type fakeOrderClient struct {
	placeFn func(ctx context.Context, in *orderv1.PlaceOrderRequest, opts ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error)
	getFn   func(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
}

func (f *fakeOrderClient) CreateReservation(context.Context, *orderv1.CreateReservationRequest, ...grpc.CallOption) (*orderv1.CreateReservationResponse, error) {
	return nil, nil
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

func TestCheckout_OK_MultiItem(t *testing.T) {
	uid, pidA, pidB, idem := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	oid := uuid.New()
	client := &fakeOrderClient{
		placeFn: func(_ context.Context, in *orderv1.PlaceOrderRequest, _ ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error) {
			if in.IdempotencyKey != idem.String() || in.UserId != uid.String() {
				t.Errorf("idem/uid wrong: %+v", in)
			}
			if len(in.Items) != 2 || in.Items[0].ProductId != pidA.String() || in.Items[1].ProductId != pidB.String() {
				t.Errorf("items wrong: %+v", in.Items)
			}
			return &orderv1.PlaceOrderResponse{Order: &orderv1.Order{
				Id: oid.String(), Status: "paid",
				Total: &orderv1.Money{Amount: "22.50", Currency: "USD"},
			}}, nil
		},
	}
	h := NewCheckoutHandlers(client)
	body, _ := json.Marshal(map[string]any{"items": []map[string]any{
		{"product_id": pidA.String(), "quantity": 2},
		{"product_id": pidB.String(), "quantity": 1},
	}})
	r := httptest.NewRequest("POST", "/checkout", bytes.NewReader(body))
	r.Header.Set("Idempotency-Key", idem.String())
	w := httptest.NewRecorder()
	h.Checkout(w, authed(uid, r))
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestCheckout_MissingIdempotencyKey_BadRequest(t *testing.T) {
	h := NewCheckoutHandlers(&fakeOrderClient{})
	body, _ := json.Marshal(map[string]any{"items": []map[string]any{
		{"product_id": uuid.New().String(), "quantity": 1},
	}})
	r := httptest.NewRequest("POST", "/checkout", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Checkout(w, authed(uuid.New(), r))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d", w.Code)
	}
}

func TestCheckout_EmptyItems_BadRequest(t *testing.T) {
	h := NewCheckoutHandlers(&fakeOrderClient{})
	body, _ := json.Marshal(map[string]any{"items": []map[string]any{}})
	r := httptest.NewRequest("POST", "/checkout", bytes.NewReader(body))
	r.Header.Set("Idempotency-Key", uuid.New().String())
	w := httptest.NewRecorder()
	h.Checkout(w, authed(uuid.New(), r))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d", w.Code)
	}
}

func TestCheckout_Unauthorized_WhenNoUserCtx(t *testing.T) {
	h := NewCheckoutHandlers(&fakeOrderClient{})
	body, _ := json.Marshal(map[string]any{"items": []map[string]any{
		{"product_id": uuid.New().String(), "quantity": 1},
	}})
	r := httptest.NewRequest("POST", "/checkout", bytes.NewReader(body))
	r.Header.Set("Idempotency-Key", uuid.New().String())
	w := httptest.NewRecorder()
	h.Checkout(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d", w.Code)
	}
}

func TestCheckout_InsufficientInventory_409(t *testing.T) {
	client := &fakeOrderClient{
		placeFn: func(context.Context, *orderv1.PlaceOrderRequest, ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error) {
			return nil, status.Error(codes.FailedPrecondition, "insufficient inventory")
		},
	}
	h := NewCheckoutHandlers(client)
	body, _ := json.Marshal(map[string]any{"items": []map[string]any{
		{"product_id": uuid.New().String(), "quantity": 1},
	}})
	r := httptest.NewRequest("POST", "/checkout", bytes.NewReader(body))
	r.Header.Set("Idempotency-Key", uuid.New().String())
	w := httptest.NewRecorder()
	h.Checkout(w, authed(uuid.New(), r))
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d", w.Code)
	}
}
