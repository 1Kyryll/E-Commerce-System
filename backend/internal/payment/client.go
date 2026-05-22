// Package payment defines the payment-provider contract used by the order
// service to charge a customer during checkout. The in-process FakeClient
// (fake.go) is used for local dev, tests, and the MVP deployment; a real
// provider (Stripe, etc.) can be slotted in later by implementing Client.
package payment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	// ErrDeclined is returned for a hard failure (card declined, insufficient
	// funds, fraud block). The caller should release the reservation.
	ErrDeclined = errors.New("payment declined")
	// ErrTimeout is reserved for the network-timeout / unknown-state branch
	// of the docs/system-design failure flow. Not produced by the fake yet —
	// added so callers can already switch on it.
	ErrTimeout = errors.New("payment timeout")
	// ErrInvalidRequest signals a malformed ChargeRequest (zero/negative
	// amount, missing currency). Callers must NOT release the reservation
	// or retry as a transient failure — fix the input.
	ErrInvalidRequest = errors.New("invalid charge request")
)

type ChargeRequest struct {
	IdempotencyKey uuid.UUID
	Amount         decimal.Decimal
	Currency       string
}

type ChargeResult struct {
	ProviderRef string
}

type Client interface {
	Charge(ctx context.Context, req ChargeRequest) (ChargeResult, error)
}
