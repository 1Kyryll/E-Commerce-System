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

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
)

func TestGetOrder_OK_OwnedByUser(t *testing.T) {
	uid, oid := uuid.New(), uuid.New()
	client := &fakeOrderClient{
		getFn: func(_ context.Context, in *orderv1.GetOrderRequest, _ ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
			if in.Id != oid.String() {
				t.Errorf("id = %s", in.Id)
			}
			return &orderv1.GetOrderResponse{Order: &orderv1.Order{
				Id: oid.String(), UserId: uid.String(), Status: "paid",
				Total: &orderv1.Money{Amount: "0", Currency: "USD"},
			}}, nil
		},
	}
	h := NewOrderHandlers(client)
	r := httptest.NewRequest("GET", "/orders/"+oid.String(), nil)
	r.SetPathValue("id", oid.String())
	w := httptest.NewRecorder()
	h.Get(w, authed(uid, r))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
}

func TestGetOrder_NotOwned_404(t *testing.T) {
	owner, other, oid := uuid.New(), uuid.New(), uuid.New()
	client := &fakeOrderClient{
		getFn: func(context.Context, *orderv1.GetOrderRequest, ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
			return &orderv1.GetOrderResponse{Order: &orderv1.Order{
				Id: oid.String(), UserId: owner.String(), Status: "paid",
				Total: &orderv1.Money{Amount: "0", Currency: "USD"},
			}}, nil
		},
	}
	h := NewOrderHandlers(client)
	r := httptest.NewRequest("GET", "/orders/"+oid.String(), nil)
	r.SetPathValue("id", oid.String())
	w := httptest.NewRecorder()
	h.Get(w, authed(other, r))
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetOrder_UpstreamNotFound_404(t *testing.T) {
	client := &fakeOrderClient{
		getFn: func(context.Context, *orderv1.GetOrderRequest, ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
			return nil, status.Error(codes.NotFound, "order not found")
		},
	}
	h := NewOrderHandlers(client)
	r := httptest.NewRequest("GET", "/orders/"+uuid.New().String(), nil)
	r.SetPathValue("id", uuid.New().String())
	w := httptest.NewRecorder()
	h.Get(w, authed(uuid.New(), r))
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d", w.Code)
	}
}

func TestGetOrder_Unauthorized_WhenNoUserCtx(t *testing.T) {
	h := NewOrderHandlers(&fakeOrderClient{})
	r := httptest.NewRequest("GET", "/orders/"+uuid.New().String(), nil)
	r.SetPathValue("id", uuid.New().String())
	w := httptest.NewRecorder()
	h.Get(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d", w.Code)
	}
}
