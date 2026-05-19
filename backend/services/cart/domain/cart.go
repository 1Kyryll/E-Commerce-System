package domain

import (
	"time"

	"github.com/google/uuid"
)

// Cart is the user-facing cart aggregate. ID is uuid.Nil when no cart row
// exists yet for the user (i.e. GetCart returned an empty cart).
type Cart struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Items  []CartItem
}

// CartItem is a single line in a Cart. Quantity is always > 0 (a line drops
// when its quantity would become zero — see Service.RemoveItem semantics).
type CartItem struct {
	ProductID uuid.UUID
	Quantity  int32
	AddedAt   time.Time
}
