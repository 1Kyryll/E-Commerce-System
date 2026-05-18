package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// userRepo is the slice of *Repo the service depends on. Defined here as an
// interface so tests can inject a fake repo without standing up a database.
type userRepo interface {
	Insert(ctx context.Context, u User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
}

// Service orchestrates signup, login, session verification, and user lookup.
// It carries the JWT signing material; transport-layer concerns (cookies,
// JSON encoding) belong to handlers.
type Service struct {
	repo   userRepo
	secret []byte
	ttl    time.Duration
}

func NewService(repo userRepo, secret []byte, ttl time.Duration) *Service {
	return &Service{repo: repo, secret: secret, ttl: ttl}
}

// TTL returns the configured session lifetime so handlers can set matching
// cookie MaxAge.
func (s *Service) TTL() time.Duration { return s.ttl }

func (s *Service) Signup(ctx context.Context, name, email, password string) (User, string, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" || password == "" {
		return User{}, "", errors.New("name, email, and password are required")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return User{}, "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.Insert(ctx, User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		return User{}, "", err
	}

	token, err := IssueToken(s.secret, user.ID, s.ttl)
	if err != nil {
		return User{}, "", fmt.Errorf("issue token: %w", err)
	}
	return user, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// Always surface ErrInvalidCredentials — never let "user not found"
		// leak as a different status than "bad password".
		return User{}, "", ErrInvalidCredentials
	}
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return User{}, "", ErrInvalidCredentials
	}
	token, err := IssueToken(s.secret, user.ID, s.ttl)
	if err != nil {
		return User{}, "", fmt.Errorf("issue token: %w", err)
	}
	return user, token, nil
}

// VerifySession parses raw, validates the signature/expiry, and returns the
// associated user id. Implements middleware.Verifier.
func (s *Service) VerifySession(raw string) (uuid.UUID, error) {
	return VerifyToken(s.secret, raw)
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (User, error) {
	return s.repo.GetByID(ctx, id)
}
