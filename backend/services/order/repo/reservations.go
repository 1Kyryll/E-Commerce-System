package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	orderdb "github.com/1Kyryll/ecommerce-demo/backend/gen/db/order"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
)

// Repo wraps the sqlc-generated orderdb.Queries with domain-friendly methods
// and pg-error mapping.
type Repo struct {
	queries *orderdb.Queries
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{queries: orderdb.New(pool)}
}

// CreateReservationParams bundles the input to the atomic CTE. The service
// is responsible for assigning a new ID via uuid.New().
type CreateReservationParams struct {
	ID             uuid.UUID
	IdempotencyKey uuid.UUID
	UserID         uuid.UUID
	ProductID      uuid.UUID
	Quantity       int32
}

// CreateReservation runs the atomic decrement-and-insert CTE behind a
// pre-check that disambiguates the three failure modes the CTE alone cannot
// tell apart (all of them collapse to pgx.ErrNoRows):
//
//  1. idempotency replay  -> ExistingReservationID is non-nil; we fetch and
//                            return the existing reservation.
//  2. product missing     -> ProductExists is false; we return ErrProductNotFound.
//  3. insufficient stock  -> only reachable case where the CTE returns ErrNoRows.
//
// The 23505 / 23503 branches remain as race-condition fallbacks: a concurrent
// caller could create with the same key after our pre-check, or the product
// could be deleted between the check and the CTE.
func (r *Repo) CreateReservation(ctx context.Context, p CreateReservationParams) (domain.Reservation, error) {
	pre, err := r.queries.CheckReservationPreconditions(ctx, orderdb.CheckReservationPreconditionsParams{
		ID:             p.ProductID,
		IdempotencyKey: p.IdempotencyKey,
	})
	if err != nil {
		return domain.Reservation{}, fmt.Errorf("check reservation preconditions: %w", err)
	}
	if pre.ExistingReservationID != uuid.Nil {
		return r.GetByIdempotencyKey(ctx, p.IdempotencyKey)
	}
	if !pre.ProductExists {
		return domain.Reservation{}, domain.ErrProductNotFound
	}

	row, err := r.queries.DecrementInventoryAndCreateReservation(ctx, orderdb.DecrementInventoryAndCreateReservationParams{
		Quantity:       p.Quantity,
		ProductID:      p.ProductID,
		ID:             p.ID,
		IdempotencyKey: p.IdempotencyKey,
		UserID:         p.UserID,
	})
	if err == nil {
		return decRowToReservation(row), nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Reservation{}, domain.ErrInsufficientInventory
	}

	// 23505: race where a concurrent caller created with the same
	// idempotency_key between our pre-check and the CTE. Refetch and return
	// the existing reservation so the API stays idempotent under contention.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return r.GetByIdempotencyKey(ctx, p.IdempotencyKey)
	}

	return domain.Reservation{}, fmt.Errorf("create reservation: %w", err)
}

func (r *Repo) GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (domain.Reservation, error) {
	row, err := r.queries.GetReservationByIdempotencyKey(ctx, key)
	if err != nil {
		return domain.Reservation{}, fmt.Errorf("get by idempotency key: %w", err)
	}
	return sharedRowToReservation(row), nil
}

func (r *Repo) ListExpired(ctx context.Context, limit int32) ([]domain.Reservation, error) {
	rows, err := r.queries.ListExpiredActiveReservations(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list expired: %w", err)
	}
	out := make([]domain.Reservation, 0, len(rows))
	for _, row := range rows {
		out = append(out, expiredRowToReservation(row))
	}
	return out, nil
}

func (r *Repo) Release(ctx context.Context, reservationID uuid.UUID) error {
	if err := r.queries.ReleaseReservationAndRestoreInventory(ctx, reservationID); err != nil {
		return fmt.Errorf("release reservation: %w", err)
	}
	return nil
}

// decRowToReservation converts the DecrementInventoryAndCreateReservation result.
func decRowToReservation(row orderdb.DecrementInventoryAndCreateReservationRow) domain.Reservation {
	return domain.Reservation{
		ID:             row.ID,
		IdempotencyKey: row.IdempotencyKey,
		ProductID:      row.ProductID,
		UserID:         row.UserID,
		Quantity:       row.Quantity,
		Status:         domain.ReservationStatus(row.Status),
		ExpiresAt:      row.ExpiresAt.Time,
		CreatedAt:      row.CreatedAt.Time,
	}
}

// sharedRowToReservation converts the shared orderdb.Reservation model
// returned by GetReservationByIdempotencyKey.
func sharedRowToReservation(row orderdb.Reservation) domain.Reservation {
	return domain.Reservation{
		ID:             row.ID,
		IdempotencyKey: row.IdempotencyKey,
		ProductID:      row.ProductID,
		UserID:         row.UserID,
		Quantity:       row.Quantity,
		Status:         domain.ReservationStatus(row.Status),
		ExpiresAt:      row.ExpiresAt.Time,
		CreatedAt:      row.CreatedAt.Time,
	}
}

// expiredRowToReservation converts the ListExpiredActiveReservations result.
// IdempotencyKey and CreatedAt are not selected by that query.
func expiredRowToReservation(row orderdb.ListExpiredActiveReservationsRow) domain.Reservation {
	return domain.Reservation{
		ID:        row.ID,
		ProductID: row.ProductID,
		UserID:    row.UserID,
		Quantity:  row.Quantity,
		Status:    domain.ReservationActive,
		ExpiresAt: row.ExpiresAt.Time,
	}
}
