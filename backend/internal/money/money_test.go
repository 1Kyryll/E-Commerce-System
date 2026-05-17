package money

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestAdd_SameCurrency(t *testing.T) {
	a := Money{Amount: decimal.RequireFromString("10.0000"), Currency: "EUR"}
	b := Money{Amount: decimal.RequireFromString("2.5000"), Currency: "EUR"}

	got, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	want := decimal.RequireFromString("12.5000")
	if !got.Amount.Equal(want) {
		t.Errorf("Amount = %s, want %s", got.Amount, want)
	}
	if got.Currency != "EUR" {
		t.Errorf("Currency = %s, want EUR", got.Currency)
	}
}

func TestAdd_CurrencyMismatch(t *testing.T) {
	a := Money{Amount: decimal.NewFromInt(1), Currency: "EUR"}
	b := Money{Amount: decimal.NewFromInt(1), Currency: "USD"}

	_, err := a.Add(b)
	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Errorf("err = %v, want ErrCurrencyMismatch", err)
	}
}

func TestMul_PreservesCurrency(t *testing.T) {
	a := Money{Amount: decimal.RequireFromString("19.99"), Currency: "EUR"}
	got := a.Mul(decimal.NewFromInt(3))
	want := decimal.RequireFromString("59.97")
	if !got.Amount.Equal(want) {
		t.Errorf("Amount = %s, want %s", got.Amount, want)
	}
	if got.Currency != "EUR" {
		t.Errorf("Currency = %s, want EUR", got.Currency)
	}
}
