package handler

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	catalogv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/catalog/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/services/catalog/domain"
)

type Service interface {
	ListProducts(ctx context.Context, cursor domain.Cursor, pageSize int32) (domain.ListResult, error)
	GetProduct(ctx context.Context, id uuid.UUID) (domain.Product, error)
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListProducts(ctx context.Context, req *catalogv1.ListProductsRequest) (*catalogv1.ListProductsResponse, error) {
	cursor, err := domain.DecodeCursor(req.GetPageCursor())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid page cursor")
	}

	result, err := h.svc.ListProducts(ctx, cursor, req.GetPageSize())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list products: %v", err)
	}

	pbProducts := make([]*catalogv1.Product, 0, len(result.Products))
	for _, p := range result.Products {
		pbProducts = append(pbProducts, productToPb(p))
	}

	return &catalogv1.ListProductsResponse{
		Products:       pbProducts,
		NextPageCursor: result.Next.Encode(),
	}, nil
}

func (h *Handler) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	p, err := h.svc.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "get product: %v", err)
	}

	return &catalogv1.GetProductResponse{Product: productToPb(p)}, nil
}

func productToPb(p domain.Product) *catalogv1.Product {
	return &catalogv1.Product{
		Id:   p.ID.String(),
		Name: p.Name,
		Price: &catalogv1.Money{
			Amount:   p.Price.Amount.StringFixed(4),
			Currency: p.Price.Currency,
		},
		InventoryAvailable: p.InventoryAvailable,
		CreatedAt:          timestamppb.New(p.CreatedAt),
		UpdatedAt:          timestamppb.New(p.UpdatedAt),
	}
}
