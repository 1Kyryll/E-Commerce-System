package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/payment"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/repo"
)

type Repo interface {
	CreateReservation(ctx context.Context, p repo.CreateReservationParams) (domain.Reservation, error)
	GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (domain.Reservation, error)
	ListExpired(ctx context.Context, limit int32) ([]domain.Reservation, error)
	Release(ctx context.Context, id uuid.UUID) error

	GetProductPriceForOrder(ctx context.Context, productID uuid.UUID) (repo.ProductPrice, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (domain.Order, error)
	GetOrderByIdempotencyKey(ctx context.Context, key uuid.UUID) (domain.Order, error)
	FinalizeOrder(ctx context.Context, p repo.FinalizeOrderParams) (domain.Order, error)
}

type Service struct {
	repo    Repo
	payment payment.Client
}

func NewService(r Repo, p payment.Client) *Service {
	return &Service{repo: r, payment: p}
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
