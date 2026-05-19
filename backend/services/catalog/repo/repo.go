package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	catalogdb "github.com/1Kyryll/ecommerce-demo/backend/gen/db/catalog"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/money"
	"github.com/1Kyryll/ecommerce-demo/backend/services/catalog/domain"
)

type Repo struct {
	queries *catalogdb.Queries
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{queries: catalogdb.New(pool)}
}

func (r *Repo) ListByCursor(ctx context.Context, cursor domain.Cursor, limit int32) ([]domain.Product, error) {
	rows, err := r.queries.ListProductsByCursor(ctx, catalogdb.ListProductsByCursorParams{
		Column1: pgtype.Timestamptz{Time: cursor.CreatedAt, Valid: true},
		Column2: cursor.ID,
		Limit:   limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	out := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		out = append(out, rowToProduct(row))
	}
	return out, nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	row, err := r.queries.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrNotFound
		}
		return domain.Product{}, fmt.Errorf("get product: %w", err)
	}
	return rowToProduct(row), nil
}

func rowToProduct(row catalogdb.Product) domain.Product {
	return domain.Product{
		ID:                 row.ID,
		Name:               row.Name,
		Price:              pgTypeNumericToMoney(row.PriceAmount, row.PriceCurrency),
		InventoryTotal:     row.InventoryTotal,
		InventoryAvailable: row.InventoryAvailable,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

func pgTypeNumericToMoney(amount pgtype.Numeric, currency string) money.Money {
	var dec decimal.Decimal
	if amount.Valid && amount.Int != nil {
		dec = decimal.NewFromBigInt(amount.Int, amount.Exp)
	} else {
		dec = decimal.Zero
	}
	return money.Money{Amount: dec, Currency: currency}
}
