package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/services/catalog/domain"
)

const (
	defaultPageSize int32 = 20
	maxPageSize     int32 = 100
)

var maxUUID = uuid.UUID{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
}

type Repo interface {
	ListByCursor(ctx context.Context, cursor domain.Cursor, limit int32) ([]domain.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProducts(ctx context.Context, cursor domain.Cursor, pageSize int32) (domain.ListResult, error) {
	pageSize = clampPageSize(pageSize)

	if cursor.IsZero() {
		cursor = domain.Cursor{
			CreatedAt: time.Now().Add(time.Hour),
			ID:        maxUUID,
		}
	}

	rows, err := s.repo.ListByCursor(ctx, cursor, pageSize+1)
	if err != nil {
		return domain.ListResult{}, err
	}

	if int32(len(rows)) <= pageSize {
		return domain.ListResult{Products: rows}, nil
	}

	rows = rows[:pageSize]
	last := rows[len(rows)-1]
	return domain.ListResult{
		Products: rows,
		Next:     domain.Cursor{CreatedAt: last.CreatedAt, ID: last.ID},
	}, nil
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func clampPageSize(n int32) int32 {
	if n <= 0 {
		return defaultPageSize
	}
	if n > maxPageSize {
		return maxPageSize
	}
	return n
}
