package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Sentinel errors returned by the service. Handlers map these to HTTP status
// codes; treat anything else as a 500-class internal error.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

// User is the in-memory representation of a row in the users table.
type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// Claims is the JWT payload. We only carry the user id (in Subject) and the
// standard issued-at / expires-at. Anything else (roles, scopes) goes into a
// future revision.
type Claims struct {
	jwt.RegisteredClaims
}
