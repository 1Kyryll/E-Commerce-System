package clients

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cartv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/cart/v1"
)

// CartClient wraps the generated gRPC client with the connection so callers
// can Close() it on shutdown.
type CartClient struct {
	conn   *grpc.ClientConn
	Client cartv1.CartServiceClient
}

// DialCart opens a plaintext gRPC connection to addr. Mirrors DialCatalog.
func DialCart(addr string) (*CartClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial cart %s: %w", addr, err)
	}
	return &CartClient{conn: conn, Client: cartv1.NewCartServiceClient(conn)}, nil
}

func (c *CartClient) Close() error {
	return c.conn.Close()
}

// SetItemQuantity sets the quantity of a cart line to an absolute value in a single atomic UPDATE.
func (c *CartClient) SetItemQuantity(ctx context.Context, userID, productID string, qty int32) (*cartv1.SetItemQuantityResponse, error) {
	return c.Client.SetItemQuantity(ctx, &cartv1.SetItemQuantityRequest{
		UserId:    userID,
		ProductId: productID,
		Quantity:  qty,
	})
}
