package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/money"
	"github.com/1Kyryll/ecommerce-demo/backend/services/catalog/domain"
)

type fakeRepo struct {
	listFn func(ctx context.Context, cursor domain.Cursor, limit int32) ([]domain.Product, error)
	getFn  func(ctx context.Context, id uuid.UUID) (domain.Product, error)
}

func (f *fakeRepo) ListByCursor(ctx context.Context, cursor domain.Cursor, limit int32) ([]domain.Product, error) {
	return f.listFn(ctx, cursor, limit)
}
func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	return f.getFn(ctx, id)
}

func buildProducts(n int) []domain.Product {
	out := make([]domain.Product, n)
	base := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		out[i] = domain.Product{
			ID:        uuid.New(),
			Name:      "p",
			Price:     money.Money{Amount: decimal.NewFromInt(1), Currency: "EUR"},
			CreatedAt: base.Add(time.Duration(-i) * time.Hour),
		}
	}
	return out
}

func TestListProducts_FirstPage_ClampsDefault(t *testing.T) {
	var seenLimit int32
	repo := &fakeRepo{
		listFn: func(_ context.Context, _ domain.Cursor, limit int32) ([]domain.Product, error) {
			seenLimit = limit
			return buildProducts(5), nil
		},
	}
	svc := NewService(repo)
	_, err := svc.ListProducts(context.Background(), domain.Cursor{}, 0)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if seenLimit != 21 {
		t.Errorf("repo limit = %d, want 21", seenLimit)
	}
}

func TestListProducts_FirstPage_UsesSentinel(t *testing.T) {
	var seenCursor domain.Cursor
	repo := &fakeRepo{
		listFn: func(_ context.Context, cursor domain.Cursor, _ int32) ([]domain.Product, error) {
			seenCursor = cursor
			return nil, nil
		},
	}
	svc := NewService(repo)
	_, err := svc.ListProducts(context.Background(), domain.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if seenCursor.IsZero() {
		t.Error("service did not substitute a non-zero sentinel for first-page cursor")
	}
	if seenCursor.CreatedAt.Before(time.Now()) {
		t.Error("sentinel CreatedAt is in the past; cursor WHERE would skip recent rows")
	}
}

func TestListProducts_ClampsAboveMax(t *testing.T) {
	var seenLimit int32
	repo := &fakeRepo{
		listFn: func(_ context.Context, _ domain.Cursor, limit int32) ([]domain.Product, error) {
			seenLimit = limit
			return nil, nil
		},
	}
	svc := NewService(repo)
	_, _ = svc.ListProducts(context.Background(), domain.Cursor{}, 5000)
	if seenLimit != 101 {
		t.Errorf("repo limit = %d, want 101 (max 100 + 1)", seenLimit)
	}
}

func TestListProducts_NextCursor_WhenHasMore(t *testing.T) {
	repo := &fakeRepo{
		listFn: func(_ context.Context, _ domain.Cursor, _ int32) ([]domain.Product, error) {
			return buildProducts(11), nil
		},
	}
	svc := NewService(repo)
	result, err := svc.ListProducts(context.Background(), domain.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(result.Products) != 10 {
		t.Errorf("len = %d, want 10", len(result.Products))
	}
	if result.Next.IsZero() {
		t.Error("Next is zero but more pages exist")
	}
	if result.Next.ID != result.Products[9].ID {
		t.Errorf("Next.ID = %v, want last product id %v", result.Next.ID, result.Products[9].ID)
	}
}

func TestListProducts_NoNextCursor_WhenAtEnd(t *testing.T) {
	repo := &fakeRepo{
		listFn: func(_ context.Context, _ domain.Cursor, _ int32) ([]domain.Product, error) {
			return buildProducts(7), nil
		},
	}
	svc := NewService(repo)
	result, err := svc.ListProducts(context.Background(), domain.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(result.Products) != 7 {
		t.Errorf("len = %d, want 7", len(result.Products))
	}
	if !result.Next.IsZero() {
		t.Errorf("Next = %+v, want zero", result.Next)
	}
}

func TestListProducts_RepoError_BubblesUp(t *testing.T) {
	repo := &fakeRepo{
		listFn: func(context.Context, domain.Cursor, int32) ([]domain.Product, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewService(repo)
	_, err := svc.ListProducts(context.Background(), domain.Cursor{}, 10)
	if err == nil {
		t.Fatal("expected error to bubble up")
	}
}

func TestGetProduct_HappyPath(t *testing.T) {
	want := domain.Product{ID: uuid.New(), Name: "Foo"}
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (domain.Product, error) { return want, nil },
	}
	svc := NewService(repo)
	got, err := svc.GetProduct(context.Background(), want.ID)
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %v", got.ID)
	}
}

func TestGetProduct_NotFound_PassesThrough(t *testing.T) {
	repo := &fakeRepo{
		getFn: func(context.Context, uuid.UUID) (domain.Product, error) {
			return domain.Product{}, domain.ErrNotFound
		},
	}
	svc := NewService(repo)
	_, err := svc.GetProduct(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
