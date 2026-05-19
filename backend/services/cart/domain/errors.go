package domain

import "errors"

// ErrProductNotFound is returned by Repo when AddItem tries to insert a
// cart_items row with a product_id that violates the FK to products. Mapped
// to gRPC NotFound by the handler.
var ErrProductNotFound = errors.New("product not found")

// ErrInvalidQuantity is returned by Service when AddItem receives a quantity
// less than 1. Mapped to gRPC InvalidArgument by the handler.
var ErrInvalidQuantity = errors.New("quantity must be positive")
