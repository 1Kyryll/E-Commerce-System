package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/money"
)

type Product struct {
	ID                 uuid.UUID
	Name               string
	Price              money.Money
	InventoryTotal     int32
	InventoryAvailable int32
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ListResult struct {
	Products []Product
	Next     Cursor
}
