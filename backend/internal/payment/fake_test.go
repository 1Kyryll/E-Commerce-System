package payment

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// newDeterministicFake returns a FakeClient with the standard 20% decline
// rate but a fixed-seed RNG so tests are reproducible.
func newDeterministicFake(seed int64) *FakeClient {
	c := NewFakeClient()
	c.rng = rand.New(rand.NewSource(seed))
	return c
}

func TestFakeClient_DefaultDeclineRate_IsTwentyPercent(t *testing.T) {
	c := NewFakeClient()
	if c.DeclineRate != 0.20 {
		t.Fatalf("DeclineRate = %v, want 0.20", c.DeclineRate)
	}
}

func TestFakeClient_Charge_ZeroDeclineRate_AlwaysSucceeds(t *testing.T) {
	c := newDeterministicFake(1)
	c.DeclineRate = 0
	for i := 0; i < 50; i++ {
		res, err := c.Charge(context.Background(), ChargeRequest{
			IdempotencyKey: uuid.New(),
			Amount:         decimal.RequireFromString("19.99"),
			Currency:       "USD",
		})
		if err != nil {
			t.Fatalf("Charge %d: %v", i, err)
		}
		if res.ProviderRef == "" {
			t.Errorf("ProviderRef empty")
		}
	}
}

func TestFakeClient_Charge_FullDeclineRate_AlwaysFails(t *testing.T) {
	c := newDeterministicFake(1)
	c.DeclineRate = 1
	_, err := c.Charge(context.Background(), ChargeRequest{
		IdempotencyKey: uuid.New(),
		Amount:         decimal.RequireFromString("1.00"),
		Currency:       "USD",
	})
	if !errors.Is(err, ErrDeclined) {
		t.Errorf("err = %v, want ErrDeclined", err)
	}
}

func TestFakeClient_Charge_DefaultRate_AboutTwentyPercentDeclines(t *testing.T) {
	c := newDeterministicFake(42) // fixed seed = reproducible distribution
	const n = 2000
	declined := 0
	for i := 0; i < n; i++ {
		_, err := c.Charge(context.Background(), ChargeRequest{
			IdempotencyKey: uuid.New(),
			Amount:         decimal.RequireFromString("9.99"),
			Currency:       "USD",
		})
		if errors.Is(err, ErrDeclined) {
			declined++
		} else if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	}
	rate := float64(declined) / float64(n)
	if rate < 0.16 || rate > 0.24 {
		t.Errorf("decline rate = %.3f over %d trials, want ~0.20 (16%%–24%%)", rate, n)
	}
}

func TestFakeClient_Charge_ZeroAmount_Rejected(t *testing.T) {
	c := newDeterministicFake(1)
	c.DeclineRate = 0 // isolate the validation branch
	_, err := c.Charge(context.Background(), ChargeRequest{
		IdempotencyKey: uuid.New(),
		Amount:         decimal.Zero,
		Currency:       "USD",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("err = %v, want ErrInvalidRequest", err)
	}
}

func TestFakeClient_Charge_MissingCurrency_Rejected(t *testing.T) {
	c := newDeterministicFake(1)
	c.DeclineRate = 0
	_, err := c.Charge(context.Background(), ChargeRequest{
		IdempotencyKey: uuid.New(),
		Amount:         decimal.RequireFromString("1.00"),
		Currency:       "",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("err = %v, want ErrInvalidRequest", err)
	}
}
