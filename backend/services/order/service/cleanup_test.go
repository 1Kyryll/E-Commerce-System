package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
)

func TestCleanupExpired_ReleasesAll(t *testing.T) {
	r1, r2 := uuid.New(), uuid.New()
	released := []uuid.UUID{}
	r := &fakeRepo{
		listFn: func(_ context.Context, _ int32) ([]domain.Reservation, error) {
			return []domain.Reservation{{ID: r1}, {ID: r2}}, nil
		},
		releaseFn: func(_ context.Context, id uuid.UUID) error {
			released = append(released, id)
			return nil
		},
	}
	svc := NewService(r)
	n, err := svc.CleanupExpired(context.Background(), 100)
	if err != nil {
		t.Fatalf("CleanupExpired: %v", err)
	}
	if n != 2 {
		t.Errorf("released count = %d, want 2", n)
	}
	if len(released) != 2 || released[0] != r1 || released[1] != r2 {
		t.Errorf("released = %v", released)
	}
}

func TestCleanupExpired_EmptyList(t *testing.T) {
	r := &fakeRepo{
		listFn: func(context.Context, int32) ([]domain.Reservation, error) { return nil, nil },
	}
	svc := NewService(r)
	n, err := svc.CleanupExpired(context.Background(), 100)
	if err != nil {
		t.Fatalf("CleanupExpired: %v", err)
	}
	if n != 0 {
		t.Errorf("n = %d, want 0", n)
	}
}

func TestCleanupExpired_ContinuesAfterReleaseFailure(t *testing.T) {
	r1, r2, r3 := uuid.New(), uuid.New(), uuid.New()
	releasedSuccessfully := 0
	r := &fakeRepo{
		listFn: func(context.Context, int32) ([]domain.Reservation, error) {
			return []domain.Reservation{{ID: r1}, {ID: r2}, {ID: r3}}, nil
		},
		releaseFn: func(_ context.Context, id uuid.UUID) error {
			if id == r2 {
				return errors.New("transient")
			}
			releasedSuccessfully++
			return nil
		},
	}
	svc := NewService(r)
	n, err := svc.CleanupExpired(context.Background(), 100)
	if err != nil {
		t.Fatalf("CleanupExpired returned error: %v", err)
	}
	if n != 2 {
		t.Errorf("released count = %d, want 2 (skipping r2 which failed)", n)
	}
	if releasedSuccessfully != 2 {
		t.Errorf("releasedSuccessfully = %d, want 2", releasedSuccessfully)
	}
}

func TestCleanupExpired_ListError_Returns(t *testing.T) {
	r := &fakeRepo{
		listFn: func(context.Context, int32) ([]domain.Reservation, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewService(r)
	_, err := svc.CleanupExpired(context.Background(), 100)
	if err == nil {
		t.Fatal("expected error to bubble up")
	}
}
