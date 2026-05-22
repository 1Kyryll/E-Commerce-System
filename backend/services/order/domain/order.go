package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderFailed    OrderStatus = "failed"
	OrderCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID             uuid.UUID
	IdempotencyKey uuid.UUID
	UserID         uuid.UUID
	TotalAmount    decimal.Decimal
	TotalCurrency  string
	Status         OrderStatus
	Items          []OrderItem
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OrderItem struct {
	ID                uuid.UUID
	OrderID           uuid.UUID
	ProductID         uuid.UUID
	ReservationID     uuid.UUID
	Quantity          int32
	UnitPriceAmount   decimal.Decimal
	UnitPriceCurrency string
}
