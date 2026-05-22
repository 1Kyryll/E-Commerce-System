package handler

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/domain"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/service"
)

// Service is the slice of *service.Service the handler depends on.
type Service interface {
	CreateReservation(ctx context.Context, idempotencyKey, userID, productID uuid.UUID, quantity int32) (domain.Reservation, error)
	PlaceOrder(ctx context.Context, idempotencyKey, userID uuid.UUID, items []service.CheckoutItem) (domain.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (domain.Order, error)
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateReservation(ctx context.Context, req *orderv1.CreateReservationRequest) (*orderv1.CreateReservationResponse, error) {
	idem, err := uuid.Parse(req.GetIdempotencyKey())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid idempotency_key")
	}
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	pid, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	res, err := h.svc.CreateReservation(ctx, idem, uid, pid, req.GetQuantity())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
		case errors.Is(err, domain.ErrInsufficientInventory):
			return nil, status.Error(codes.FailedPrecondition, "insufficient inventory")
		case errors.Is(err, domain.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product not found")
		default:
			return nil, status.Errorf(codes.Internal, "create reservation: %v", err)
		}
	}

	return &orderv1.CreateReservationResponse{Reservation: reservationToPb(res)}, nil
}

func (h *Handler) PlaceOrder(ctx context.Context, req *orderv1.PlaceOrderRequest) (*orderv1.PlaceOrderResponse, error) {
	idem, err := uuid.Parse(req.GetIdempotencyKey())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid idempotency_key")
	}
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	items := make([]service.CheckoutItem, 0, len(req.GetItems()))
	for i, it := range req.GetItems() {
		pid, err := uuid.Parse(it.GetProductId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "items[%d]: invalid product_id", i)
		}
		items = append(items, service.CheckoutItem{ProductID: pid, Quantity: it.GetQuantity()})
	}

	order, err := h.svc.PlaceOrder(ctx, idem, uid, items)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmptyCheckout):
			return nil, status.Error(codes.InvalidArgument, "checkout has no items")
		case errors.Is(err, domain.ErrDuplicateProduct):
			return nil, status.Error(codes.InvalidArgument, "duplicate product in checkout")
		case errors.Is(err, domain.ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
		case errors.Is(err, domain.ErrInsufficientInventory):
			return nil, status.Error(codes.FailedPrecondition, "insufficient inventory")
		case errors.Is(err, domain.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product not found")
		case errors.Is(err, domain.ErrPaymentDeclined):
			return nil, status.Error(codes.FailedPrecondition, "payment declined")
		case errors.Is(err, domain.ErrReservationNotActive):
			return nil, status.Error(codes.FailedPrecondition, "reservation not active")
		default:
			return nil, status.Errorf(codes.Internal, "place order: %v", err)
		}
	}
	return &orderv1.PlaceOrderResponse{Order: orderToPb(order)}, nil
}

func (h *Handler) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	order, err := h.svc.GetOrder(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			return nil, status.Error(codes.NotFound, "order not found")
		default:
			return nil, status.Errorf(codes.Internal, "get order: %v", err)
		}
	}
	return &orderv1.GetOrderResponse{Order: orderToPb(order)}, nil
}

func reservationToPb(r domain.Reservation) *orderv1.Reservation {
	return &orderv1.Reservation{
		Id:             r.ID.String(),
		IdempotencyKey: r.IdempotencyKey.String(),
		ProductId:      r.ProductID.String(),
		UserId:         r.UserID.String(),
		Quantity:       r.Quantity,
		Status:         string(r.Status),
		ExpiresAt:      timestamppb.New(r.ExpiresAt),
		CreatedAt:      timestamppb.New(r.CreatedAt),
	}
}

func orderToPb(o domain.Order) *orderv1.Order {
	items := make([]*orderv1.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, &orderv1.OrderItem{
			ProductId: it.ProductID.String(),
			Quantity:  it.Quantity,
			UnitPrice: &orderv1.Money{
				Amount:   it.UnitPriceAmount.String(),
				Currency: it.UnitPriceCurrency,
			},
		})
	}
	return &orderv1.Order{
		Id:             o.ID.String(),
		IdempotencyKey: o.IdempotencyKey.String(),
		UserId:         o.UserID.String(),
		Total: &orderv1.Money{
			Amount:   o.TotalAmount.String(),
			Currency: o.TotalCurrency,
		},
		Status:    string(o.Status),
		Items:     items,
		CreatedAt: timestamppb.New(o.CreatedAt),
	}
}
