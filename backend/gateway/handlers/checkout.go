package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

type orderClient interface {
	CreateReservation(ctx context.Context, in *orderv1.CreateReservationRequest, opts ...grpc.CallOption) (*orderv1.CreateReservationResponse, error)
	PlaceOrder(ctx context.Context, in *orderv1.PlaceOrderRequest, opts ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error)
	GetOrder(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
}

type CheckoutHandlers struct {
	order orderClient
}

func NewCheckoutHandlers(order orderClient) *CheckoutHandlers {
	return &CheckoutHandlers{order: order}
}

type checkoutItemJSON struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type checkoutRequest struct {
	Items []checkoutItemJSON `json:"items"`
}

type orderItemJSON struct {
	ProductID string           `json:"product_id"`
	Quantity  int32            `json:"quantity"`
	UnitPrice productMoneyJSON `json:"unit_price"`
}

type orderJSON struct {
	ID             string           `json:"id"`
	IdempotencyKey string           `json:"idempotency_key"`
	UserID         string           `json:"user_id"`
	Total          productMoneyJSON `json:"total"`
	Status         string           `json:"status"`
	Items          []orderItemJSON  `json:"items"`
	CreatedAt      string           `json:"created_at,omitempty"`
}

func (h *CheckoutHandlers) Checkout(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idem := r.Header.Get("Idempotency-Key")
	if idem == "" {
		http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
		return
	}
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "items must be non-empty", http.StatusBadRequest)
		return
	}
	items := make([]*orderv1.CheckoutItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, &orderv1.CheckoutItem{ProductId: it.ProductID, Quantity: it.Quantity})
	}
	resp, err := h.order.PlaceOrder(r.Context(), &orderv1.PlaceOrderRequest{
		IdempotencyKey: idem,
		UserId:         uid.String(),
		Items:          items,
	})
	if err != nil {
		writeCheckoutError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, pbToOrderJSON(resp.GetOrder()))
}

// writeCheckoutError maps the FailedPrecondition cases (insufficient inventory,
// payment declined, reservation not active) to 409 Conflict, which is the
// closest HTTP semantic for "the request was valid but the world said no".
func writeCheckoutError(w http.ResponseWriter, err error) {
	if status.Code(err) == codes.FailedPrecondition {
		http.Error(w, "conflict: "+status.Convert(err).Message(), http.StatusConflict)
		return
	}
	writeUpstreamError(w, err)
}

func pbToOrderJSON(o *orderv1.Order) orderJSON {
	items := make([]orderItemJSON, 0, len(o.GetItems()))
	for _, it := range o.GetItems() {
		items = append(items, orderItemJSON{
			ProductID: it.GetProductId(),
			Quantity:  it.GetQuantity(),
			UnitPrice: productMoneyJSON{
				Amount:   it.GetUnitPrice().GetAmount(),
				Currency: it.GetUnitPrice().GetCurrency(),
			},
		})
	}
	out := orderJSON{
		ID:             o.GetId(),
		IdempotencyKey: o.GetIdempotencyKey(),
		UserID:         o.GetUserId(),
		Total: productMoneyJSON{
			Amount:   o.GetTotal().GetAmount(),
			Currency: o.GetTotal().GetCurrency(),
		},
		Status: o.GetStatus(),
		Items:  items,
	}
	if t := o.GetCreatedAt(); t != nil {
		out.CreatedAt = t.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	return out
}
