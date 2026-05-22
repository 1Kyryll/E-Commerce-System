package clients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
)

type OrderClient struct {
	conn   *grpc.ClientConn
	Client orderv1.OrderServiceClient
}

func DialOrder(addr string) (*OrderClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial order %s: %w", addr, err)
	}
	return &OrderClient{conn: conn, Client: orderv1.NewOrderServiceClient(conn)}, nil
}

func (c *OrderClient) Close() error { return c.conn.Close() }
