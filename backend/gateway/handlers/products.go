package handlers

import (
	"context"
	"net/http"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/catalog/v1"
)

// catalogClient is the slice of catalogv1.CatalogServiceClient this package
// uses. Defined as an interface so tests can substitute a fake.
type catalogClient interface {
	ListProducts(ctx context.Context, in *catalogv1.ListProductsRequest, opts ...grpc.CallOption) (*catalogv1.ListProductsResponse, error)
	GetProduct(ctx context.Context, in *catalogv1.GetProductRequest, opts ...grpc.CallOption) (*catalogv1.GetProductResponse, error)
}

type ProductHandlers struct {
	catalog catalogClient
}

func NewProductHandlers(catalog catalogClient) *ProductHandlers {
	return &ProductHandlers{catalog: catalog}
}

// JSON shapes — independent of the gRPC proto types so the REST contract can
// evolve without touching the wire format.
type productMoneyJSON struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type productJSON struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	Price              productMoneyJSON `json:"price"`
	InventoryAvailable int32            `json:"inventory_available"`
	CreatedAt          string           `json:"created_at,omitempty"`
	UpdatedAt          string           `json:"updated_at,omitempty"`
}

type productListJSON struct {
	Products       []productJSON `json:"products"`
	NextPageCursor string        `json:"next_page_cursor"`
}

func (h *ProductHandlers) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pageCursor := q.Get("cursor")

	var pageSize int32
	if s := q.Get("page_size"); s != "" {
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			http.Error(w, "page_size must be an integer", http.StatusBadRequest)
			return
		}
		pageSize = int32(n)
	}

	resp, err := h.catalog.ListProducts(r.Context(), &catalogv1.ListProductsRequest{
		PageCursor: pageCursor,
		PageSize:   pageSize,
	})
	if err != nil {
		writeUpstreamError(w, err)
		return
	}

	out := productListJSON{
		Products:       make([]productJSON, 0, len(resp.GetProducts())),
		NextPageCursor: resp.GetNextPageCursor(),
	}
	for _, p := range resp.GetProducts() {
		out.Products = append(out.Products, pbToProductJSON(p))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ProductHandlers) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp, err := h.catalog.GetProduct(r.Context(), &catalogv1.GetProductRequest{Id: id})
	if err != nil {
		writeUpstreamError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pbToProductJSON(resp.GetProduct()))
}

func pbToProductJSON(p *catalogv1.Product) productJSON {
	out := productJSON{
		ID:                 p.GetId(),
		Name:               p.GetName(),
		InventoryAvailable: p.GetInventoryAvailable(),
		Price: productMoneyJSON{
			Amount:   p.GetPrice().GetAmount(),
			Currency: p.GetPrice().GetCurrency(),
		},
	}
	if t := p.GetCreatedAt(); t != nil {
		out.CreatedAt = t.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	if t := p.GetUpdatedAt(); t != nil {
		out.UpdatedAt = t.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	return out
}

// writeUpstreamError maps gRPC status codes from upstream to HTTP. Anything
// without an explicit mapping becomes 502 (gateway succeeded, upstream did not).
func writeUpstreamError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.NotFound:
		http.Error(w, "not found", http.StatusNotFound)
	case codes.InvalidArgument:
		http.Error(w, "bad request", http.StatusBadRequest)
	case codes.Unauthenticated:
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	case codes.PermissionDenied:
		http.Error(w, "forbidden", http.StatusForbidden)
	default:
		http.Error(w, "upstream error", http.StatusBadGateway)
	}
}
