package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc"

	cartv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/cart/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/middleware"
)

// cartClient is the slice of cartv1.CartServiceClient this package uses.
// Defined as an interface so tests can substitute a fake.
type cartClient interface {
	GetCart(ctx context.Context, in *cartv1.GetCartRequest, opts ...grpc.CallOption) (*cartv1.GetCartResponse, error)
	AddItem(ctx context.Context, in *cartv1.AddItemRequest, opts ...grpc.CallOption) (*cartv1.AddItemResponse, error)
	RemoveItem(ctx context.Context, in *cartv1.RemoveItemRequest, opts ...grpc.CallOption) (*cartv1.RemoveItemResponse, error)
	ClearCart(ctx context.Context, in *cartv1.ClearCartRequest, opts ...grpc.CallOption) (*cartv1.ClearCartResponse, error)
	SetItemQuantity(ctx context.Context, in *cartv1.SetItemQuantityRequest, opts ...grpc.CallOption) (*cartv1.SetItemQuantityResponse, error)
}

type CartHandlers struct {
	cart cartClient
}

func NewCartHandlers(cart cartClient) *CartHandlers {
	return &CartHandlers{cart: cart}
}

type cartItemJSON struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	AddedAt   string `json:"added_at,omitempty"`
}

type cartJSON struct {
	ID     string         `json:"id"`
	UserID string         `json:"user_id"`
	Items  []cartItemJSON `json:"items"`
}

type addItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type setItemQuantityRequest struct {
	Quantity int32 `json:"quantity"`
}

func (h *CartHandlers) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	resp, err := h.cart.GetCart(r.Context(), &cartv1.GetCartRequest{UserId: uid.String()})
	if err != nil {
		writeUpstreamError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pbToCartJSON(resp.GetCart()))
}

func (h *CartHandlers) AddItem(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req addItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	resp, err := h.cart.AddItem(r.Context(), &cartv1.AddItemRequest{
		UserId:    uid.String(),
		ProductId: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		writeUpstreamError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pbToCartJSON(resp.GetCart()))
}

func (h *CartHandlers) RemoveItem(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	productID := r.PathValue("product_id")
	resp, err := h.cart.RemoveItem(r.Context(), &cartv1.RemoveItemRequest{
		UserId: uid.String(), ProductId: productID,
	})
	if err != nil {
		writeUpstreamError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pbToCartJSON(resp.GetCart()))
}

func (h *CartHandlers) Clear(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := h.cart.ClearCart(r.Context(), &cartv1.ClearCartRequest{UserId: uid.String()}); err != nil {
		writeUpstreamError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandlers) SetItemQuantity(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	productID := r.PathValue("product_id")
	var req setItemQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Quantity < 0 {
		http.Error(w, "quantity must be > 0", http.StatusBadRequest)
		return
	}

	if _, err := h.cart.SetItemQuantity(r.Context(), &cartv1.SetItemQuantityRequest{UserId: uid.String(), ProductId: productID, Quantity: req.Quantity}); err != nil {
		writeUpstreamError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pbToCartJSON(c *cartv1.Cart) cartJSON {
	out := cartJSON{
		ID:     c.GetId(),
		UserID: c.GetUserId(),
		Items:  make([]cartItemJSON, 0, len(c.GetItems())),
	}
	for _, it := range c.GetItems() {
		item := cartItemJSON{ProductID: it.GetProductId(), Quantity: it.GetQuantity()}
		if t := it.GetAddedAt(); t != nil {
			item.AddedAt = t.AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
		out.Items = append(out.Items, item)
	}
	return out
}
