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

	cartv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/cart/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

type fakeCartClient struct {
	getFn    func(ctx context.Context, in *cartv1.GetCartRequest, opts ...grpc.CallOption) (*cartv1.GetCartResponse, error)
	addFn    func(ctx context.Context, in *cartv1.AddItemRequest, opts ...grpc.CallOption) (*cartv1.AddItemResponse, error)
	removeFn func(ctx context.Context, in *cartv1.RemoveItemRequest, opts ...grpc.CallOption) (*cartv1.RemoveItemResponse, error)
	clearFn  func(ctx context.Context, in *cartv1.ClearCartRequest, opts ...grpc.CallOption) (*cartv1.ClearCartResponse, error)
	setQtyFn func(ctx context.Context, in *cartv1.SetItemQuantityRequest, opts ...grpc.CallOption) (*cartv1.SetItemQuantityResponse, error)
}

func (f *fakeCartClient) GetCart(ctx context.Context, in *cartv1.GetCartRequest, opts ...grpc.CallOption) (*cartv1.GetCartResponse, error) {
	return f.getFn(ctx, in, opts...)
}
func (f *fakeCartClient) AddItem(ctx context.Context, in *cartv1.AddItemRequest, opts ...grpc.CallOption) (*cartv1.AddItemResponse, error) {
	return f.addFn(ctx, in, opts...)
}
func (f *fakeCartClient) RemoveItem(ctx context.Context, in *cartv1.RemoveItemRequest, opts ...grpc.CallOption) (*cartv1.RemoveItemResponse, error) {
	return f.removeFn(ctx, in, opts...)
}
func (f *fakeCartClient) ClearCart(ctx context.Context, in *cartv1.ClearCartRequest, opts ...grpc.CallOption) (*cartv1.ClearCartResponse, error) {
	return f.clearFn(ctx, in, opts...)
}
func (f *fakeCartClient) SetItemQuantity(ctx context.Context, in *cartv1.SetItemQuantityRequest, opts ...grpc.CallOption) (*cartv1.SetItemQuantityResponse, error) {
	return f.setQtyFn(ctx, in, opts...)
}

func withUserID(req *http.Request, uid uuid.UUID) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey(), uid))
}

func TestCartGet_OK(t *testing.T) {
	uid, pid := uuid.New(), uuid.New()
	client := &fakeCartClient{
		getFn: func(_ context.Context, in *cartv1.GetCartRequest, _ ...grpc.CallOption) (*cartv1.GetCartResponse, error) {
			if in.GetUserId() != uid.String() {
				t.Errorf("user_id = %q", in.GetUserId())
			}
			return &cartv1.GetCartResponse{Cart: &cartv1.Cart{
				Id: uuid.New().String(), UserId: uid.String(),
				Items: []*cartv1.CartItem{{ProductId: pid.String(), Quantity: 2}},
			}}, nil
		},
	}
	h := NewCartHandlers(client)
	req := withUserID(httptest.NewRequest(http.MethodGet, "/cart", nil), uid)
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body struct {
		UserID string `json:"user_id"`
		Items  []struct {
			ProductID string `json:"product_id"`
			Quantity  int32  `json:"quantity"`
		} `json:"items"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.UserID != uid.String() || len(body.Items) != 1 {
		t.Errorf("body = %+v", body)
	}
}

func TestCartGet_NoUserCtx_Unauthorized(t *testing.T) {
	h := NewCartHandlers(&fakeCartClient{})
	req := httptest.NewRequest(http.MethodGet, "/cart", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestCartAdd_OK(t *testing.T) {
	uid, pid := uuid.New(), uuid.New()
	client := &fakeCartClient{
		addFn: func(_ context.Context, in *cartv1.AddItemRequest, _ ...grpc.CallOption) (*cartv1.AddItemResponse, error) {
			if in.GetUserId() != uid.String() || in.GetProductId() != pid.String() || in.GetQuantity() != 3 {
				t.Errorf("req = %+v", in)
			}
			return &cartv1.AddItemResponse{Cart: &cartv1.Cart{Id: uuid.New().String(), UserId: uid.String()}}, nil
		},
	}
	h := NewCartHandlers(client)

	body, _ := json.Marshal(map[string]any{"product_id": pid.String(), "quantity": 3})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewReader(body)), uid)
	rec := httptest.NewRecorder()
	h.AddItem(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestCartAdd_BadJSON_BadRequest(t *testing.T) {
	uid := uuid.New()
	h := NewCartHandlers(&fakeCartClient{})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewReader([]byte("{not json"))), uid)
	rec := httptest.NewRecorder()
	h.AddItem(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestCartAdd_UpstreamNotFound_NotFound(t *testing.T) {
	uid := uuid.New()
	client := &fakeCartClient{
		addFn: func(context.Context, *cartv1.AddItemRequest, ...grpc.CallOption) (*cartv1.AddItemResponse, error) {
			return nil, status.Error(codes.NotFound, "product missing")
		},
	}
	h := NewCartHandlers(client)
	body, _ := json.Marshal(map[string]any{"product_id": uuid.New().String(), "quantity": 1})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewReader(body)), uid)
	rec := httptest.NewRecorder()
	h.AddItem(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestCartAdd_UpstreamInvalidArgument_BadRequest(t *testing.T) {
	uid := uuid.New()
	client := &fakeCartClient{
		addFn: func(context.Context, *cartv1.AddItemRequest, ...grpc.CallOption) (*cartv1.AddItemResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "bad qty")
		},
	}
	h := NewCartHandlers(client)
	body, _ := json.Marshal(map[string]any{"product_id": uuid.New().String(), "quantity": 0})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewReader(body)), uid)
	rec := httptest.NewRecorder()
	h.AddItem(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestCartRemove_OK(t *testing.T) {
	uid, pid := uuid.New(), uuid.New()
	client := &fakeCartClient{
		removeFn: func(_ context.Context, in *cartv1.RemoveItemRequest, _ ...grpc.CallOption) (*cartv1.RemoveItemResponse, error) {
			if in.GetUserId() != uid.String() || in.GetProductId() != pid.String() {
				t.Errorf("req = %+v", in)
			}
			return &cartv1.RemoveItemResponse{Cart: &cartv1.Cart{Id: uuid.New().String(), UserId: uid.String()}}, nil
		},
	}
	h := NewCartHandlers(client)
	req := withUserID(httptest.NewRequest(http.MethodDelete, "/cart/items/"+pid.String(), nil), uid)
	req.SetPathValue("product_id", pid.String())
	rec := httptest.NewRecorder()
	h.RemoveItem(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestCartClear_OK(t *testing.T) {
	uid := uuid.New()
	called := false
	client := &fakeCartClient{
		clearFn: func(_ context.Context, in *cartv1.ClearCartRequest, _ ...grpc.CallOption) (*cartv1.ClearCartResponse, error) {
			called = true
			if in.GetUserId() != uid.String() {
				t.Errorf("user_id = %q", in.GetUserId())
			}
			return &cartv1.ClearCartResponse{}, nil
		},
	}
	h := NewCartHandlers(client)
	req := withUserID(httptest.NewRequest(http.MethodDelete, "/cart", nil), uid)
	rec := httptest.NewRecorder()
	h.Clear(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d", rec.Code)
	}
	if !called {
		t.Error("Clear was not called")
	}
}
