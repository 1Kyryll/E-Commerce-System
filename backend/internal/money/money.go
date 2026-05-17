package money

import (
	"errors"

	"github.com/shopspring/decimal"
)

var ErrCurrencyMismatch = errors.New("currency mismatch")

type Money struct {
	Amount   decimal.Decimal
	Currency string // ISO 4217
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount.Add(other.Amount), Currency: m.Currency}, nil
}

func (m Money) Mul(factor decimal.Decimal) Money {
	return Money{Amount: m.Amount.Mul(factor), Currency: m.Currency}
}
