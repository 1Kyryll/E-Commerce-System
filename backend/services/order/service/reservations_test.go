package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/repo"
)

type fakeRepo struct {
	createFn  func(ctx context.Context, p repo.CreateReservationParams) (domain.Reservation, error)
	getFn     func(ctx context.Context, key uuid.UUID) (domain.Reservation, error)
	listFn    func(ctx context.Context, limit int32) ([]domain.Reservation, error)
	releaseFn func(ctx context.Context, id uuid.UUID) error
}

func (f *fakeRepo) CreateReservation(ctx context.Context, p repo.CreateReservationParams) (domain.Reservation, error) {
	return f.createFn(ctx, p)
}
func (f *fakeRepo) GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (domain.Reservation, error) {
	return f.getFn(ctx, key)
}
func (f *fakeRepo) ListExpired(ctx context.Context, limit int32) ([]domain.Reservation, error) {
	return f.listFn(ctx, limit)
}
func (f *fakeRepo) Release(ctx context.Context, id uuid.UUID) error {
	return f.releaseFn(ctx, id)
}

func TestCreateReservation_HappyPath(t *testing.T) {
	idem, uid, pid := uuid.New(), uuid.New(), uuid.New()
	var gotParams repo.CreateReservationParams
	r := &fakeRepo{
		createFn: func(_ context.Context, p repo.CreateReservationParams) (domain.Reservation, error) {
			gotParams = p
			return domain.Reservation{
				ID: p.ID, IdempotencyKey: p.IdempotencyKey,
				ProductID: p.ProductID, UserID: p.UserID,
				Quantity: p.Quantity, Status: domain.ReservationActive,
				ExpiresAt: time.Now().Add(15 * time.Minute),
			}, nil
		},
	}
	svc := NewService(r)
	res, err := svc.CreateReservation(context.Background(), idem, uid, pid, 2)
	if err != nil {
		t.Fatalf("CreateReservation: %v", err)
	}
	if res.Quantity != 2 || res.Status != domain.ReservationActive {
		t.Errorf("res = %+v", res)
	}
	if gotParams.IdempotencyKey != idem || gotParams.UserID != uid || gotParams.ProductID != pid || gotParams.Quantity != 2 {
		t.Errorf("repo got %+v", gotParams)
	}
	if gotParams.ID == uuid.Nil {
		t.Error("service did not assign a new reservation ID")
	}
}

func TestCreateReservation_InvalidQuantity_Zero(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.CreateReservation(context.Background(), uuid.New(), uuid.New(), uuid.New(), 0)
	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v, want ErrInvalidQuantity", err)
	}
}

func TestCreateReservation_InvalidQuantity_Negative(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.CreateReservation(context.Background(), uuid.New(), uuid.New(), uuid.New(), -5)
	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("err = %v, want ErrInvalidQuantity", err)
	}
}

func TestCreateReservation_InsufficientInventory(t *testing.T) {
	r := &fakeRepo{
		createFn: func(context.Context, repo.CreateReservationParams) (domain.Reservation, error) {
			return domain.Reservation{}, domain.ErrInsufficientInventory
		},
	}
	svc := NewService(r)
	_, err := svc.CreateReservation(context.Background(), uuid.New(), uuid.New(), uuid.New(), 1)
	if !errors.Is(err, domain.ErrInsufficientInventory) {
		t.Errorf("err = %v, want ErrInsufficientInventory", err)
	}
}

func TestCreateReservation_ProductNotFound(t *testing.T) {
	r := &fakeRepo{
		createFn: func(context.Context, repo.CreateReservationParams) (domain.Reservation, error) {
			return domain.Reservation{}, domain.ErrProductNotFound
		},
	}
	svc := NewService(r)
	_, err := svc.CreateReservation(context.Background(), uuid.New(), uuid.New(), uuid.New(), 1)
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Errorf("err = %v, want ErrProductNotFound", err)
	}
}

func TestCreateReservation_IdempotencyReplay(t *testing.T) {
	existing := domain.Reservation{
		ID: uuid.New(), Quantity: 7, Status: domain.ReservationActive,
	}
	r := &fakeRepo{
		createFn: func(context.Context, repo.CreateReservationParams) (domain.Reservation, error) {
			return existing, nil
		},
	}
	svc := NewService(r)
	got, err := svc.CreateReservation(context.Background(), uuid.New(), uuid.New(), uuid.New(), 99)
	if err != nil {
		t.Fatalf("CreateReservation: %v", err)
	}
	if got.ID != existing.ID || got.Quantity != 7 {
		t.Errorf("got = %+v, want existing %+v", got, existing)
	}
}
