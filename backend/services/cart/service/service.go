package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/services/cart/domain"
)

type Repo interface {
	Get(ctx context.Context, userID uuid.UUID) (domain.Cart, error)
	AddItem(ctx context.Context, userID, productID uuid.UUID, qty int32) (domain.Cart, error)
	RemoveItem(ctx context.Context, userID, productID uuid.UUID) (domain.Cart, error)
	Clear(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	return s.repo.Get(ctx, userID)
}

func (s *Service) AddItem(ctx context.Context, userID, productID uuid.UUID, qty int32) (domain.Cart, error) {
	if qty < 1 {
		return domain.Cart{}, domain.ErrInvalidQuantity
	}
	return s.repo.AddItem(ctx, userID, productID, qty)
}

func (s *Service) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (domain.Cart, error) {
	return s.repo.RemoveItem(ctx, userID, productID)
}

func (s *Service) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.repo.Clear(ctx, userID)
}
