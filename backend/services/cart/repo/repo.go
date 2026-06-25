package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	cartdb "github.com/1Kyryll/ecommerce-demo/backend/gen/db/cart"
	"github.com/1Kyryll/ecommerce-demo/backend/services/cart/domain"
)

// Repo wraps the sqlc-generated cartdb.Queries and owns its own transactions
// for the multi-statement mutations.
type Repo struct {
	pool    *pgxpool.Pool
	queries *cartdb.Queries
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool, queries: cartdb.New(pool)}
}

// Get returns the user's cart and items. If no cart row exists yet, returns
// a zero-id Cart with UserID set and no items — NOT an error.
func (r *Repo) Get(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	cart, err := r.queries.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Cart{UserID: userID}, nil
		}
		return domain.Cart{}, fmt.Errorf("get cart: %w", err)
	}

	items, err := r.queries.ListCartItems(ctx, cart.ID)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("list cart items: %w", err)
	}

	return cartToDomain(cart, items), nil
}

// AddItem upserts the user's cart and increments the (product_id) line by
// qty. Returns the updated cart with all items.
func (r *Repo) AddItem(ctx context.Context, userID, productID uuid.UUID, qty int32) (domain.Cart, error) {
	var out domain.Cart
	err := r.inTx(ctx, func(q *cartdb.Queries) error {
		cart, err := q.UpsertCart(ctx, cartdb.UpsertCartParams{
			ID:     uuid.New(),
			UserID: userID,
		})
		if err != nil {
			return fmt.Errorf("upsert cart: %w", err)
		}

		if _, err := q.UpsertCartItem(ctx, cartdb.UpsertCartItemParams{
			ID:        uuid.New(),
			CartID:    cart.ID,
			ProductID: productID,
			Quantity:  qty,
		}); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return domain.ErrProductNotFound
			}
			return fmt.Errorf("upsert cart item: %w", err)
		}

		items, err := q.ListCartItems(ctx, cart.ID)
		if err != nil {
			return fmt.Errorf("list cart items: %w", err)
		}

		out = cartToDomain(cart, items)
		return nil
	})
	return out, err
}

// RemoveItem deletes the (product_id) line from the user's cart, if present.
// If the cart doesn't exist, returns an empty cart without error (idempotent).
func (r *Repo) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (domain.Cart, error) {
	var out domain.Cart
	err := r.inTx(ctx, func(q *cartdb.Queries) error {
		cart, err := q.GetCartByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				out = domain.Cart{UserID: userID}
				return nil
			}
			return fmt.Errorf("get cart: %w", err)
		}

		if err := q.RemoveCartItem(ctx, cartdb.RemoveCartItemParams{
			CartID:    cart.ID,
			ProductID: productID,
		}); err != nil {
			return fmt.Errorf("remove cart item: %w", err)
		}

		items, err := q.ListCartItems(ctx, cart.ID)
		if err != nil {
			return fmt.Errorf("list cart items: %w", err)
		}

		out = cartToDomain(cart, items)
		return nil
	})
	return out, err
}

// Clear removes all items from the user's cart. If the cart doesn't exist,
// it's a no-op.
func (r *Repo) Clear(ctx context.Context, userID uuid.UUID) error {
	return r.inTx(ctx, func(q *cartdb.Queries) error {
		cart, err := q.GetCartByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("get cart: %w", err)
		}
		if err := q.ClearCart(ctx, cart.ID); err != nil {
			return fmt.Errorf("clear cart: %w", err)
		}
		return nil
	})
}

// SetItemQuantity sets the (product_id) line in the user's cart to an absolute
// quantity in a single atomic UPDATE. The cart is resolved by user_id first,
// mirroring RemoveItem/Clear. Returns domain.ErrProductNotFound if the user has
// no cart or the product isn't in it. Callers must pass qty >= 1 (the service
// rejects non-positive quantities before reaching here).
func (r *Repo) SetItemQuantity(ctx context.Context, userID, productID uuid.UUID, qty int32) error {
	return r.inTx(ctx, func(q *cartdb.Queries) error {
		cart, err := q.GetCartByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrProductNotFound
			}
			return fmt.Errorf("get cart: %w", err)
		}

		rows, err := q.SetCartItemQuantity(ctx, cartdb.SetCartItemQuantityParams{
			CartID:    cart.ID,
			ProductID: productID,
			Quantity:  qty,
		})
		if err != nil {
			return fmt.Errorf("set cart item quantity: %w", err)
		}
		if rows == 0 {
			return domain.ErrProductNotFound
		}
		return nil
	})
}

// inTx opens a tx-bound *Queries, runs fn, and commits on nil. Any error
// rolls the tx back.
func (r *Repo) inTx(ctx context.Context, fn func(*cartdb.Queries) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // safe to call after Commit — no-op

	if err := fn(r.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func cartToDomain(row cartdb.Cart, items []cartdb.CartItem) domain.Cart {
	out := domain.Cart{
		ID:     row.ID,
		UserID: row.UserID,
		Items:  make([]domain.CartItem, 0, len(items)),
	}
	for _, it := range items {
		out.Items = append(out.Items, domain.CartItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			AddedAt:   it.AddedAt.Time,
		})
	}
	return out
}
