package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cartv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/cart/v1"
	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

// orderClient is the slice of orderv1.OrderServiceClient the gateway uses.
// CreateReservation is deliberately excluded — it's an order-service-internal
// concern that PlaceOrder orchestrates; the gateway never calls it directly.
type orderClient interface {
	PlaceOrder(ctx context.Context, in *orderv1.PlaceOrderRequest, opts ...grpc.CallOption) (*orderv1.PlaceOrderResponse, error)
	GetOrder(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
}

type CheckoutHandlers struct {
	order orderClient
	cart  cartClient
}

func NewCheckoutHandlers(order orderClient, cart cartClient) *CheckoutHandlers {
	return &CheckoutHandlers{order: order, cart: cart}
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

// Checkout reads the caller's cart, places an order for its contents, and on
// success clears the cart. No request body — items come from the cart, the
// only client input is the Idempotency-Key header.
//
// Retry semantics: a successful checkout clears the cart, so a retry with
// the same Idempotency-Key after the response was already received will see
// an empty cart and return 400. Mid-flight retries (before cart-clear)
// succeed because PlaceOrder is idempotent on the parent key. Clients that
// lost a response can recover via GET /orders.
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

	cartResp, err := h.cart.GetCart(r.Context(), &cartv1.GetCartRequest{UserId: uid.String()})
	if err != nil {
		writeUpstreamError(w, err)
		return
	}
	cartItems := cartResp.GetCart().GetItems()
	if len(cartItems) == 0 {
		http.Error(w, "cart is empty", http.StatusBadRequest)
		return
	}

	items := make([]*orderv1.CheckoutItem, 0, len(cartItems))
	for _, ci := range cartItems {
		items = append(items, &orderv1.CheckoutItem{
			ProductId: ci.GetProductId(),
			Quantity:  ci.GetQuantity(),
		})
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

	// Best-effort cart clear. A failure here doesn't roll the order back —
	// the order is already finalized. We log and continue so the caller
	// gets their 201; a stale cart can be cleared manually.
	if _, clearErr := h.cart.ClearCart(r.Context(), &cartv1.ClearCartRequest{UserId: uid.String()}); clearErr != nil {
		slog.WarnContext(r.Context(), "checkout: cart clear failed",
			"user_id", uid, "order_id", resp.GetOrder().GetId(), "err", clearErr)
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
