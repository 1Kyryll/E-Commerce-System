package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReservationStatus enumerates the lifecycle states from docs/system-design.md.
// Active means inventory has been decremented. Consumed and Released are
// terminal states.
type ReservationStatus string

const (
	ReservationActive   ReservationStatus = "active"
	ReservationConsumed ReservationStatus = "consumed"
	ReservationReleased ReservationStatus = "released"
)

// DefaultReservationTTL is how long an active reservation holds inventory
// before the cleanup worker releases it. Matches the INTERVAL '15 minutes'
// baked into the sqlc DecrementInventoryAndCreateReservation query.
const DefaultReservationTTL = 15 * time.Minute

// Reservation is the in-memory representation of a reservations row.
type Reservation struct {
	ID             uuid.UUID
	IdempotencyKey uuid.UUID
	ProductID      uuid.UUID
	UserID         uuid.UUID
	Quantity       int32
	Status         ReservationStatus
	ExpiresAt      time.Time
	CreatedAt      time.Time
}
