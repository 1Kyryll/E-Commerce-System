package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestOrderStatusConstants(t *testing.T) {
	if OrderPending != "pending" || OrderPaid != "paid" ||
		OrderFailed != "failed" || OrderCancelled != "cancelled" {
		t.Fatalf("status constants drifted")
	}
}

func TestOrderTotal_MatchesItem(t *testing.T) {
	o := Order{
		ID:            uuid.New(),
		TotalAmount:   decimal.RequireFromString("19.98"),
		TotalCurrency: "USD",
		Items: []OrderItem{{
			ProductID:         uuid.New(),
			Quantity:          2,
			UnitPriceAmount:   decimal.RequireFromString("9.99"),
			UnitPriceCurrency: "USD",
		}},
	}
	got := o.Items[0].UnitPriceAmount.Mul(decimal.NewFromInt(int64(o.Items[0].Quantity)))
	if !got.Equal(o.TotalAmount) {
		t.Fatalf("derived total %s != stored %s", got, o.TotalAmount)
	}
}
