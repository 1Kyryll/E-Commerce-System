package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/services/cart/domain"
)

type fakeRepo struct {
	getFn     func(ctx context.Context, userID uuid.UUID) (domain.Cart, error)
	addFn     func(ctx context.Context, userID, productID uuid.UUID, qty int32) (domain.Cart, error)
	removeFn  func(ctx context.Context, userID, productID uuid.UUID) (domain.Cart, error)
	clearFn   func(ctx context.Context, userID uuid.UUID) error
	setcartitemqtyFn func(ctx context.Context, cartID, productID uuid.UUID, qty int32) error
}

func (f *fakeRepo) Get(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	return f.getFn(ctx, userID)
}
func (f *fakeRepo) AddItem(ctx context.Context, userID, productID uuid.UUID, qty int32) (domain.Cart, error) {
	return f.addFn(ctx, userID, productID, qty)
}
func (f *fakeRepo) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (domain.Cart, error) {
	return f.removeFn(ctx, userID, productID)
}
func (f *fakeRepo) Clear(ctx context.Context, userID uuid.UUID) error {
	return f.clearFn(ctx, userID)
}
func (f *fakeRepo) SetItemQuantity(ctx context.Context, cartID, productID uuid.UUID, qty int32) error{
	return f.setcartitemqtyFn(ctx,cartID, productID, qty)
}

func TestAddItem_Happy(t *testing.T) {
	uid, pid := uuid.New(), uuid.New()
	want := domain.Cart{ID: uuid.New(), UserID: uid, Items: []domain.CartItem{{ProductID: pid, Quantity: 2, AddedAt: time.Now()}}}
	repo := &fakeRepo{
		addFn: func(_ context.Context, u, p uuid.UUID, q int32) (domain.Cart, error) {
			if u != uid || p != pid || q != 2 {
				t.Errorf("repo got (%s, %s, %d)", u, p, q)
			}
			return want, nil
		},
	}
	svc := NewService(repo)
	got, err := svc.AddItem(context.Background(), uid, pid, 2)
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if got.ID != want.ID || len(got.Items) != 1 {
		t.Errorf("got = %+v", got)
	}
}

func TestAddItem_InvalidQuantity_Zero(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.AddItem(context.Background(), uuid.New(), uuid.New(), 0)
	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v, want ErrInvalidQuantity", err)
	}
}

func TestAddItem_InvalidQuantity_Negative(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.AddItem(context.Background(), uuid.New(), uuid.New(), -3)
	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v, want ErrInvalidQuantity", err)
	}
}

func TestAddItem_ProductNotFound_BubblesUp(t *testing.T) {
	repo := &fakeRepo{
		addFn: func(context.Context, uuid.UUID, uuid.UUID, int32) (domain.Cart, error) {
			return domain.Cart{}, domain.ErrProductNotFound
		},
	}
	svc := NewService(repo)
	_, err := svc.AddItem(context.Background(), uuid.New(), uuid.New(), 1)
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Errorf("err = %v, want ErrProductNotFound", err)
	}
}

func TestRemoveItem_Happy(t *testing.T) {
	uid, pid := uuid.New(), uuid.New()
	want := domain.Cart{ID: uuid.New(), UserID: uid}
	repo := &fakeRepo{
		removeFn: func(_ context.Context, u, p uuid.UUID) (domain.Cart, error) {
			if u != uid || p != pid {
				t.Errorf("repo got (%s, %s)", u, p)
			}
			return want, nil
		},
	}
	svc := NewService(repo)
	got, err := svc.RemoveItem(context.Background(), uid, pid)
	if err != nil {
		t.Fatalf("RemoveItem: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("got.ID = %v", got.ID)
	}
}

func TestClearCart_Happy(t *testing.T) {
	called := false
	uid := uuid.New()
	repo := &fakeRepo{
		clearFn: func(_ context.Context, u uuid.UUID) error {
			called = true
			if u != uid {
				t.Errorf("user = %s", u)
			}
			return nil
		},
	}
	svc := NewService(repo)
	if err := svc.ClearCart(context.Background(), uid); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("repo.Clear was not called")
	}
}

func TestGetCart_Happy(t *testing.T) {
	uid := uuid.New()
	want := domain.Cart{UserID: uid, ID: uuid.New(), Items: []domain.CartItem{{ProductID: uuid.New(), Quantity: 3}}}
	repo := &fakeRepo{
		getFn: func(_ context.Context, u uuid.UUID) (domain.Cart, error) {
			if u != uid {
				t.Errorf("user = %s", u)
			}
			return want, nil
		},
	}
	svc := NewService(repo)
	got, err := svc.GetCart(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || len(got.Items) != 1 {
		t.Errorf("got = %+v", got)
	}
}

func TestGetCart_EmptyForNewUser(t *testing.T) {
	uid := uuid.New()
	repo := &fakeRepo{
		getFn: func(_ context.Context, u uuid.UUID) (domain.Cart, error) {
			return domain.Cart{UserID: u}, nil
		},
	}
	svc := NewService(repo)
	got, err := svc.GetCart(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != uuid.Nil {
		t.Errorf("got.ID = %v, want uuid.Nil for empty cart", got.ID)
	}
	if len(got.Items) != 0 {
		t.Errorf("Items len = %d, want 0", len(got.Items))
	}
}

func TestSetItemQuantity_Happy(t *testing.T) {
	cartID, pid := uuid.New(), uuid.New()
	repo := &fakeRepo{
		setcartitemqtyFn: func(_ context.Context, c, p uuid.UUID, q int32) error {
			if c != cartID || p != pid || q != 5 {
				t.Errorf("repo got (%s, %s, %d)", c, p, q)
			}
			return nil
		},
	}
	svc := NewService(repo)
	if err := svc.SetItemQuantity(context.Background(), cartID, pid, 5); err != nil {
		t.Fatalf("SetItemQuantity: %v", err)
	}
}

func TestSetItemQuantity_InvalidQuantity(t *testing.T) {
	svc := NewService(&fakeRepo{})
	err := svc.SetItemQuantity(context.Background(), uuid.New(), uuid.New(), 0)
	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v, want ErrInvalidQuantity", err)
	}
	err = svc.SetItemQuantity(context.Background(), uuid.New(), uuid.New(), -1)
	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v, want ErrInvalidQuantity", err)
	}
}
