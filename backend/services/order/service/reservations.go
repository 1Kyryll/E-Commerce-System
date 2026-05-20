package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/repo"
)

type Repo interface {
	CreateReservation(ctx context.Context, p repo.CreateReservationParams) (domain.Reservation, error)
	GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (domain.Reservation, error)
	ListExpired(ctx context.Context, limit int32) ([]domain.Reservation, error)
	Release(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateReservation(ctx context.Context, idempotencyKey, userID, productID uuid.UUID, quantity int32) (domain.Reservation, error) {
	if quantity < 1 {
		return domain.Reservation{}, domain.ErrInvalidQuantity
	}
	return s.repo.CreateReservation(ctx, repo.CreateReservationParams{
		ID:             uuid.New(),
		IdempotencyKey: idempotencyKey,
		UserID:         userID,
		ProductID:      productID,
		Quantity:       quantity,
	})
}
