package payment

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// DefaultDeclineRate is the share of charges the FakeClient rejects with
// ErrDeclined by default. We deliberately exercise the failure path in
// load tests and dashboards instead of running a happy-only fake.
const DefaultDeclineRate = 0.20

// FakeClient is the in-process Client used outside of real payment runs.
// DeclineRate (0..1) controls the probability of returning ErrDeclined on
// any given Charge; tests override `rng` for reproducibility.
type FakeClient struct {
	DeclineRate float64

	mu  sync.Mutex
	rng *rand.Rand
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		DeclineRate: DefaultDeclineRate,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *FakeClient) Charge(_ context.Context, req ChargeRequest) (ChargeResult, error) {
	if req.Amount.IsZero() || req.Amount.IsNegative() {
		return ChargeResult{}, fmt.Errorf("amount must be positive: %w", ErrInvalidRequest)
	}
	if req.Currency == "" {
		return ChargeResult{}, fmt.Errorf("currency required: %w", ErrInvalidRequest)
	}

	c.mu.Lock()
	roll := c.rng.Float64()
	c.mu.Unlock()

	if roll < c.DeclineRate {
		return ChargeResult{}, ErrDeclined
	}
	return ChargeResult{ProviderRef: fmt.Sprintf("fake_%s", req.IdempotencyKey)}, nil
}
