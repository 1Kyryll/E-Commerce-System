package clients

import (
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	catalogv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/catalog/v1"
)

// CatalogClient wraps the generated gRPC client with the underlying connection
// so the caller can Close() it on shutdown.
type CatalogClient struct {
	conn   *grpc.ClientConn
	Client catalogv1.CatalogServiceClient
}

// DialCatalog opens a plaintext gRPC connection to addr (e.g. "localhost:9000").
// In production this would use TLS; for local dev and same-host docker-compose
// deployment, insecure is fine.
func DialCatalog(addr string) (*CatalogClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial catalog %s: %w", addr, err)
	}
	return &CatalogClient{conn: conn, Client: catalogv1.NewCatalogServiceClient(conn)}, nil
}

func (c *CatalogClient) Close() error {
	return c.conn.Close()
}
