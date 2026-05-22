package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	orderdb "github.com/1Kyryll/ecommerce-demo/backend/gen/db/order"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
)

// ProductPrice is the lightweight slice of a product row that the order
// service needs to compute totals at checkout time.
type ProductPrice struct {
	Amount   decimal.Decimal
	Currency string
}

func (r *Repo) GetProductPriceForOrder(ctx context.Context, productID uuid.UUID) (ProductPrice, error) {
	row, err := r.queries.GetProductPriceForOrder(ctx, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProductPrice{}, domain.ErrProductNotFound
		}
		return ProductPrice{}, fmt.Errorf("get product price: %w", err)
	}
	return ProductPrice{Amount: numericToDecimal(row.PriceAmount), Currency: row.PriceCurrency}, nil
}

func (r *Repo) GetOrderByID(ctx context.Context, id uuid.UUID) (domain.Order, error) {
	row, err := r.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}
	items, err := r.queries.ListOrderItems(ctx, id)
	if err != nil {
		return domain.Order{}, fmt.Errorf("list order items: %w", err)
	}
	out := domain.Order{
		ID:             row.ID,
		IdempotencyKey: row.IdempotencyKey,
		UserID:         row.UserID,
		TotalAmount:    numericToDecimal(row.TotalAmount),
		TotalCurrency:  row.TotalCurrency,
		Status:         domain.OrderStatus(row.Status),
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		Items:          make([]domain.OrderItem, 0, len(items)),
	}
	for _, it := range items {
		oi := domain.OrderItem{
			ID:                it.ID,
			OrderID:           it.OrderID,
			ProductID:         it.ProductID,
			Quantity:          it.Quantity,
			UnitPriceAmount:   numericToDecimal(it.UnitPriceAmount),
			UnitPriceCurrency: it.UnitPriceCurrency,
		}
		if it.ReservationID.Valid {
			oi.ReservationID = it.ReservationID.Bytes
		}
		out.Items = append(out.Items, oi)
	}
	return out, nil
}

func (r *Repo) GetOrderByIdempotencyKey(ctx context.Context, key uuid.UUID) (domain.Order, error) {
	row, err := r.queries.GetOrderByIdempotencyKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, fmt.Errorf("get order by key: %w", err)
	}
	return r.GetOrderByID(ctx, row.ID)
}

// FinalizeOrderParams bundles the inputs to FinalizeOrder. The service has
// already chosen UUIDs and computed the total.
type FinalizeOrderParams struct {
	OrderID           uuid.UUID
	IdempotencyKey    uuid.UUID
	UserID            uuid.UUID
	TotalAmount       decimal.Decimal
	TotalCurrency     string
	OrderItemID       uuid.UUID
	ReservationID     uuid.UUID
	ProductID         uuid.UUID
	Quantity          int32
	UnitPriceAmount   decimal.Decimal
	UnitPriceCurrency string
	OutboxEventID     uuid.UUID
}

// FinalizeOrder runs the second transaction from docs/system-design.md:
// consume the reservation, insert the order + order_item, write the outbox
// event. All four happen atomically. Returns ErrReservationNotActive if the
// reservation was no longer 'active' at the moment of consumption.
func (r *Repo) FinalizeOrder(ctx context.Context, p FinalizeOrderParams) (domain.Order, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Order{}, fmt.Errorf("begin finalize tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := r.queries.WithTx(tx)

	if _, err := q.ConsumeReservationByID(ctx, p.ReservationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrReservationNotActive
		}
		return domain.Order{}, fmt.Errorf("consume reservation: %w", err)
	}

	orderRow, err := q.InsertOrder(ctx, orderdb.InsertOrderParams{
		ID:             p.OrderID,
		IdempotencyKey: p.IdempotencyKey,
		UserID:         p.UserID,
		TotalAmount:    decimalToNumeric(p.TotalAmount),
		TotalCurrency:  p.TotalCurrency,
		Status:         string(domain.OrderPaid),
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("insert order: %w", err)
	}

	itemRow, err := q.InsertOrderItem(ctx, orderdb.InsertOrderItemParams{
		ID:                p.OrderItemID,
		OrderID:           p.OrderID,
		ProductID:         p.ProductID,
		ReservationID:     pgtype.UUID{Bytes: p.ReservationID, Valid: true},
		Quantity:          p.Quantity,
		UnitPriceAmount:   decimalToNumeric(p.UnitPriceAmount),
		UnitPriceCurrency: p.UnitPriceCurrency,
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("insert order item: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"order_id":   p.OrderID,
		"user_id":    p.UserID,
		"total":      p.TotalAmount.String(),
		"currency":   p.TotalCurrency,
		"product_id": p.ProductID,
		"quantity":   p.Quantity,
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("marshal outbox payload: %w", err)
	}

	if _, err := q.InsertOutboxEvent(ctx, orderdb.InsertOutboxEventParams{
		ID:            p.OutboxEventID,
		AggregateType: "order",
		AggregateID:   p.OrderID,
		EventType:     "OrderPlaced",
		Payload:       payload,
	}); err != nil {
		return domain.Order{}, fmt.Errorf("insert outbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("commit finalize tx: %w", err)
	}

	itemReservationID := uuid.Nil
	if itemRow.ReservationID.Valid {
		itemReservationID = itemRow.ReservationID.Bytes
	}
	return domain.Order{
		ID:             orderRow.ID,
		IdempotencyKey: orderRow.IdempotencyKey,
		UserID:         orderRow.UserID,
		TotalAmount:    numericToDecimal(orderRow.TotalAmount),
		TotalCurrency:  orderRow.TotalCurrency,
		Status:         domain.OrderStatus(orderRow.Status),
		CreatedAt:      orderRow.CreatedAt.Time,
		UpdatedAt:      orderRow.UpdatedAt.Time,
		Items: []domain.OrderItem{{
			ID:                itemRow.ID,
			OrderID:           itemRow.OrderID,
			ProductID:         itemRow.ProductID,
			ReservationID:     itemReservationID,
			Quantity:          itemRow.Quantity,
			UnitPriceAmount:   numericToDecimal(itemRow.UnitPriceAmount),
			UnitPriceCurrency: itemRow.UnitPriceCurrency,
		}},
	}, nil
}

// numericToDecimal converts a pgtype.Numeric (sqlc's representation of
// NUMERIC columns) to a shopspring/decimal.Decimal. Invalid / null values
// collapse to decimal.Zero — callers that need to distinguish null from
// zero must check Valid themselves before calling this.
func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid || n.Int == nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(n.Int, n.Exp)
}

// decimalToNumeric converts a shopspring/decimal.Decimal to pgtype.Numeric
// for write paths.
func decimalToNumeric(d decimal.Decimal) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   new(big.Int).Set(d.Coefficient()),
		Exp:   d.Exponent(),
		Valid: true,
	}
}
