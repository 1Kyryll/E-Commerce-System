package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	authdb "github.com/1Kyryll/ecommerce-demo/backend/gen/db/auth"
)

// Repo wraps the sqlc-generated authdb.Queries with domain-friendly methods.
// It translates between database types and domain types and maps no-row
// errors to ErrInvalidCredentials so callers can decide what to surface
// without leaking which lookup failed.
type Repo struct {
	queries *authdb.Queries
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{queries: authdb.New(pool)}
}

func (r *Repo) Insert(ctx context.Context, u User) (User, error) {
	row, err := r.queries.InsertUser(ctx, authdb.InsertUserParams{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailExists
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return rowToUser(row), nil
}

func (r *Repo) GetByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return rowToUser(row), nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return rowToUser(row), nil
}

func rowToUser(row authdb.User) User {
	return User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
	}
}
