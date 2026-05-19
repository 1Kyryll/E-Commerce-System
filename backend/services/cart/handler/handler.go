package handler

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	cartv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/cart/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/services/cart/domain"
)

type Service interface {
	GetCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error)
	AddItem(ctx context.Context, userID, productID uuid.UUID, qty int32) (domain.Cart, error)
	RemoveItem(ctx context.Context, userID, productID uuid.UUID) (domain.Cart, error)
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetCart(ctx context.Context, req *cartv1.GetCartRequest) (*cartv1.GetCartResponse, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	cart, err := h.svc.GetCart(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get cart: %v", err)
	}
	return &cartv1.GetCartResponse{Cart: cartToPb(cart)}, nil
}

func (h *Handler) AddItem(ctx context.Context, req *cartv1.AddItemRequest) (*cartv1.AddItemResponse, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	pid, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}
	cart, err := h.svc.AddItem(ctx, uid, pid, req.GetQuantity())
	if err != nil {
		return nil, mapServiceError(err, "add item")
	}
	return &cartv1.AddItemResponse{Cart: cartToPb(cart)}, nil
}

func (h *Handler) RemoveItem(ctx context.Context, req *cartv1.RemoveItemRequest) (*cartv1.RemoveItemResponse, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	pid, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}
	cart, err := h.svc.RemoveItem(ctx, uid, pid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "remove item: %v", err)
	}
	return &cartv1.RemoveItemResponse{Cart: cartToPb(cart)}, nil
}

func (h *Handler) ClearCart(ctx context.Context, req *cartv1.ClearCartRequest) (*cartv1.ClearCartResponse, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if err := h.svc.ClearCart(ctx, uid); err != nil {
		return nil, status.Errorf(codes.Internal, "clear cart: %v", err)
	}
	return &cartv1.ClearCartResponse{}, nil
}

func mapServiceError(err error, op string) error {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, "product not found")
	case errors.Is(err, domain.ErrInvalidQuantity):
		return status.Error(codes.InvalidArgument, "quantity must be positive")
	default:
		return status.Errorf(codes.Internal, "%s: %v", op, err)
	}
}

func cartToPb(c domain.Cart) *cartv1.Cart {
	items := make([]*cartv1.CartItem, 0, len(c.Items))
	for _, it := range c.Items {
		items = append(items, &cartv1.CartItem{
			ProductId: it.ProductID.String(),
			Quantity:  it.Quantity,
			AddedAt:   timestamppb.New(it.AddedAt),
		})
	}
	return &cartv1.Cart{
		Id:     c.ID.String(),
		UserId: c.UserID.String(),
		Items:  items,
	}
}
