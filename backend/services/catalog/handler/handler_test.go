package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/catalog/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/money"
	"github.com/1Kyryll/ecommerce-demo/backend/services/catalog/domain"
)

type fakeSvc struct {
	listFn func(ctx context.Context, cursor domain.Cursor, pageSize int32) (domain.ListResult, error)
	getFn  func(ctx context.Context, id uuid.UUID) (domain.Product, error)
}

func (f *fakeSvc) ListProducts(ctx context.Context, cursor domain.Cursor, pageSize int32) (domain.ListResult, error) {
	return f.listFn(ctx, cursor, pageSize)
}
func (f *fakeSvc) GetProduct(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	return f.getFn(ctx, id)
}

func TestListProducts_OK(t *testing.T) {
	uid := uuid.New()
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	svc := &fakeSvc{
		listFn: func(context.Context, domain.Cursor, int32) (domain.ListResult, error) {
			return domain.ListResult{
				Products: []domain.Product{{
					ID:                 uid,
					Name:               "Widget",
					Price:              money.Money{Amount: decimal.RequireFromString("19.9900"), Currency: "EUR"},
					InventoryAvailable: 5,
					CreatedAt:          now,
					UpdatedAt:          now,
				}},
				Next: domain.Cursor{CreatedAt: now, ID: uid},
			}, nil
		},
	}
	h := New(svc)
	resp, err := h.ListProducts(context.Background(), &catalogv1.ListProductsRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(resp.Products) != 1 {
		t.Fatalf("len = %d", len(resp.Products))
	}
	p := resp.Products[0]
	if p.Id != uid.String() {
		t.Errorf("id = %s", p.Id)
	}
	if p.Name != "Widget" {
		t.Errorf("name = %s", p.Name)
	}
	if p.Price.Amount != "19.9900" {
		t.Errorf("price amount = %s", p.Price.Amount)
	}
	if p.Price.Currency != "EUR" {
		t.Errorf("price currency = %s", p.Price.Currency)
	}
	if p.InventoryAvailable != 5 {
		t.Errorf("inventory_available = %d", p.InventoryAvailable)
	}
	if resp.NextPageCursor == "" {
		t.Error("next_page_cursor empty when more pages exist")
	}
}

func TestListProducts_BadCursor_InvalidArgument(t *testing.T) {
	svc := &fakeSvc{}
	h := New(svc)
	_, err := h.ListProducts(context.Background(), &catalogv1.ListProductsRequest{PageCursor: "!!not-base64!!"})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestListProducts_ServiceError_Internal(t *testing.T) {
	svc := &fakeSvc{
		listFn: func(context.Context, domain.Cursor, int32) (domain.ListResult, error) {
			return domain.ListResult{}, errors.New("db down")
		},
	}
	h := New(svc)
	_, err := h.ListProducts(context.Background(), &catalogv1.ListProductsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("code = %v, want Internal", status.Code(err))
	}
}

func TestGetProduct_OK(t *testing.T) {
	uid := uuid.New()
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	svc := &fakeSvc{
		getFn: func(context.Context, uuid.UUID) (domain.Product, error) {
			return domain.Product{
				ID:                 uid,
				Name:               "Widget",
				Price:              money.Money{Amount: decimal.RequireFromString("19.9900"), Currency: "EUR"},
				InventoryAvailable: 5,
				CreatedAt:          now,
				UpdatedAt:          now,
			}, nil
		},
	}
	h := New(svc)
	resp, err := h.GetProduct(context.Background(), &catalogv1.GetProductRequest{Id: uid.String()})
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if resp.Product.Id != uid.String() {
		t.Errorf("id = %s", resp.Product.Id)
	}
}

func TestGetProduct_BadID_InvalidArgument(t *testing.T) {
	svc := &fakeSvc{}
	h := New(svc)
	_, err := h.GetProduct(context.Background(), &catalogv1.GetProductRequest{Id: "not-a-uuid"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", status.Code(err))
	}
}

func TestGetProduct_NotFound_NotFound(t *testing.T) {
	svc := &fakeSvc{
		getFn: func(context.Context, uuid.UUID) (domain.Product, error) {
			return domain.Product{}, domain.ErrNotFound
		},
	}
	h := New(svc)
	_, err := h.GetProduct(context.Background(), &catalogv1.GetProductRequest{Id: uuid.New().String()})
	if status.Code(err) != codes.NotFound {
		t.Errorf("code = %v, want NotFound", status.Code(err))
	}
}
