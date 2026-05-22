package domain

import "errors"

// Sentinel errors. The handler maps these to gRPC status codes.
var (
	// ErrInsufficientInventory means the atomic CTE returned no rows because
	// inventory_available was less than the requested quantity. Maps to
	// gRPC FailedPrecondition.
	ErrInsufficientInventory = errors.New("insufficient inventory")

	// ErrProductNotFound is returned when the FK to products(id) is violated.
	// Maps to gRPC NotFound.
	ErrProductNotFound = errors.New("product not found")

	// ErrInvalidQuantity is returned for quantity < 1. Maps to InvalidArgument.
	ErrInvalidQuantity = errors.New("quantity must be positive")
)

// ErrOrderNotFound is returned by GetOrder when no row matches. Maps to NotFound.
var ErrOrderNotFound = errors.New("order not found")

// ErrPaymentDeclined is returned when the payment provider hard-fails the
// charge. The reservation has already been released. Maps to FailedPrecondition.
var ErrPaymentDeclined = errors.New("payment declined")

// ErrReservationNotActive means the reservation existed but is in a terminal
// state ('consumed' or 'released'), so PlaceOrder cannot finalize it. Maps to
// FailedPrecondition.
var ErrReservationNotActive = errors.New("reservation not active")
