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
)

// Service is the slice of *service.Service the handler depends on.
type Service interface {
	CreateReservation(ctx context.Context, idempotencyKey, userID, productID uuid.UUID, quantity int32) (domain.Reservation, error)
}

// Handler implements orderv1.OrderServiceServer. PlaceOrder and GetOrder are
// stubbed as Unimplemented for now — they land in a follow-up plan.
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

func (h *Handler) PlaceOrder(_ context.Context, _ *orderv1.PlaceOrderRequest) (*orderv1.PlaceOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "place_order not implemented")
}

func (h *Handler) GetOrder(_ context.Context, _ *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "get_order not implemented")
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
