package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	catalogv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/catalog/v1"
)

type fakeCatalogClient struct {
	listFn func(ctx context.Context, in *catalogv1.ListProductsRequest, opts ...grpc.CallOption) (*catalogv1.ListProductsResponse, error)
	getFn  func(ctx context.Context, in *catalogv1.GetProductRequest, opts ...grpc.CallOption) (*catalogv1.GetProductResponse, error)
}

func (f *fakeCatalogClient) ListProducts(ctx context.Context, in *catalogv1.ListProductsRequest, opts ...grpc.CallOption) (*catalogv1.ListProductsResponse, error) {
	return f.listFn(ctx, in, opts...)
}
func (f *fakeCatalogClient) GetProduct(ctx context.Context, in *catalogv1.GetProductRequest, opts ...grpc.CallOption) (*catalogv1.GetProductResponse, error) {
	return f.getFn(ctx, in, opts...)
}

func TestListProducts_OK(t *testing.T) {
	now := timestamppb.Now()
	client := &fakeCatalogClient{
		listFn: func(_ context.Context, in *catalogv1.ListProductsRequest, _ ...grpc.CallOption) (*catalogv1.ListProductsResponse, error) {
			if in.GetPageSize() != 5 {
				t.Errorf("page_size = %d, want 5", in.GetPageSize())
			}
			if in.GetPageCursor() != "cursor-token" {
				t.Errorf("page_cursor = %q", in.GetPageCursor())
			}
			return &catalogv1.ListProductsResponse{
				Products: []*catalogv1.Product{{
					Id:                 "00000000-0000-0000-0000-000000000001",
					Name:               "Widget",
					Price:              &catalogv1.Money{Amount: "19.9900", Currency: "EUR"},
					InventoryAvailable: 5,
					CreatedAt:          now,
					UpdatedAt:          now,
				}},
				NextPageCursor: "next-token",
			}, nil
		},
	}

	h := NewProductHandlers(client)
	req := httptest.NewRequest(http.MethodGet, "/products?cursor=cursor-token&page_size=5", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Products []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Price struct {
				Amount   string `json:"amount"`
				Currency string `json:"currency"`
			} `json:"price"`
			InventoryAvailable int32 `json:"inventory_available"`
		} `json:"products"`
		NextPageCursor string `json:"next_page_cursor"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Products) != 1 {
		t.Fatalf("len = %d", len(body.Products))
	}
	if body.Products[0].Name != "Widget" || body.Products[0].Price.Amount != "19.9900" {
		t.Errorf("product wrong: %+v", body.Products[0])
	}
	if body.NextPageCursor != "next-token" {
		t.Errorf("next_page_cursor = %q", body.NextPageCursor)
	}
}

func TestListProducts_DefaultsWhenQueryMissing(t *testing.T) {
	var seenSize int32
	var seenCursor string
	client := &fakeCatalogClient{
		listFn: func(_ context.Context, in *catalogv1.ListProductsRequest, _ ...grpc.CallOption) (*catalogv1.ListProductsResponse, error) {
			seenSize = in.GetPageSize()
			seenCursor = in.GetPageCursor()
			return &catalogv1.ListProductsResponse{}, nil
		},
	}
	h := NewProductHandlers(client)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if seenSize != 0 {
		t.Errorf("page_size = %d, want 0 (server clamps to default)", seenSize)
	}
	if seenCursor != "" {
		t.Errorf("cursor = %q, want \"\"", seenCursor)
	}
}

func TestListProducts_BadPageSize_BadRequest(t *testing.T) {
	h := NewProductHandlers(&fakeCatalogClient{})
	req := httptest.NewRequest(http.MethodGet, "/products?page_size=abc", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestListProducts_UpstreamInvalidArgument_BadRequest(t *testing.T) {
	client := &fakeCatalogClient{
		listFn: func(context.Context, *catalogv1.ListProductsRequest, ...grpc.CallOption) (*catalogv1.ListProductsResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "bad cursor")
		},
	}
	h := NewProductHandlers(client)
	req := httptest.NewRequest(http.MethodGet, "/products?cursor=garbage", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestListProducts_UpstreamInternal_BadGateway(t *testing.T) {
	client := &fakeCatalogClient{
		listFn: func(context.Context, *catalogv1.ListProductsRequest, ...grpc.CallOption) (*catalogv1.ListProductsResponse, error) {
			return nil, status.Error(codes.Internal, "boom")
		},
	}
	h := NewProductHandlers(client)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rec.Code)
	}
}

func TestGetProduct_OK(t *testing.T) {
	now := timestamppb.Now()
	client := &fakeCatalogClient{
		getFn: func(_ context.Context, in *catalogv1.GetProductRequest, _ ...grpc.CallOption) (*catalogv1.GetProductResponse, error) {
			if in.GetId() != "00000000-0000-0000-0000-000000000001" {
				t.Errorf("id = %q", in.GetId())
			}
			return &catalogv1.GetProductResponse{Product: &catalogv1.Product{
				Id:                 "00000000-0000-0000-0000-000000000001",
				Name:               "Widget",
				Price:              &catalogv1.Money{Amount: "19.9900", Currency: "EUR"},
				InventoryAvailable: 5,
				CreatedAt:          now,
				UpdatedAt:          now,
			}}, nil
		},
	}

	h := NewProductHandlers(client)
	req := httptest.NewRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000001", nil)
	req.SetPathValue("id", "00000000-0000-0000-0000-000000000001")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Name != "Widget" {
		t.Errorf("name = %q", body.Name)
	}
}

func TestGetProduct_UpstreamNotFound_NotFound(t *testing.T) {
	client := &fakeCatalogClient{
		getFn: func(context.Context, *catalogv1.GetProductRequest, ...grpc.CallOption) (*catalogv1.GetProductResponse, error) {
			return nil, status.Error(codes.NotFound, "missing")
		},
	}
	h := NewProductHandlers(client)
	req := httptest.NewRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000099", nil)
	req.SetPathValue("id", "00000000-0000-0000-0000-000000000099")
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestGetProduct_UpstreamInvalidArgument_BadRequest(t *testing.T) {
	client := &fakeCatalogClient{
		getFn: func(context.Context, *catalogv1.GetProductRequest, ...grpc.CallOption) (*catalogv1.GetProductResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "bad id")
		},
	}
	h := NewProductHandlers(client)
	req := httptest.NewRequest(http.MethodGet, "/products/not-a-uuid", nil)
	req.SetPathValue("id", "not-a-uuid")
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}
