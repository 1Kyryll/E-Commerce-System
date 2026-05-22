package handlers

import (
	"net/http"

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

type OrderHandlers struct {
	order orderClient
}

func NewOrderHandlers(order orderClient) *OrderHandlers {
	return &OrderHandlers{order: order}
}

func (h *OrderHandlers) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")
	resp, err := h.order.GetOrder(r.Context(), &orderv1.GetOrderRequest{Id: id})
	if err != nil {
		writeUpstreamError(w, err)
		return
	}
	// Owner-only: mask non-owner reads as 404 so we don't confirm the
	// existence of someone else's order id.
	if resp.GetOrder().GetUserId() != uid.String() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, pbToOrderJSON(resp.GetOrder()))
}
