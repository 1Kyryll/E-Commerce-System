package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
)

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (domain.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}
