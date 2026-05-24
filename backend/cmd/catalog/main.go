package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	catalogv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/catalog/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/database"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/observability"
	"github.com/1Kyryll/ecommerce-demo/backend/services/catalog/handler"
	"github.com/1Kyryll/ecommerce-demo/backend/services/catalog/repo"
	"github.com/1Kyryll/ecommerce-demo/backend/services/catalog/service"
)

func serviceName(defaultName string) string {
	if v := os.Getenv("OTEL_SERVICE_NAME"); v != "" {
		return v
	}
	return defaultName
}

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("catalog: config: %v", err)
	}

	obsShutdown, err := observability.Init(context.Background(), serviceName("catalog"))
	if err != nil {
		log.Fatalf("catalog: observability: %v", err)
	}
	defer func() {
		if err := obsShutdown(context.Background()); err != nil {
			slog.Error("observability shutdown", "err", err)
		}
	}()

	slog.SetDefault(slog.New(observability.NewSlogHandler(os.Stderr, slog.LevelDebug)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, database.PoolConfig{URL: cfg.DatabaseURL})
	if err != nil {
		log.Fatalf("catalog: db: %v", err)
	}
	defer pool.Close()

	catalogSvc := service.NewService(repo.NewRepo(pool))
	catalogHandler := handler.New(catalogSvc)

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	catalogv1.RegisterCatalogServiceServer(grpcServer, catalogHandler)
	reflection.Register(grpcServer) // enables grpcurl

	addr := fmt.Sprintf(":%d", cfg.GRPCPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("catalog: listen: %v", err)
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("catalog gRPC listening", "addr", addr)
		if err := grpcServer.Serve(ln); err != nil {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			log.Fatalf("catalog: serve: %v", err)
		}
	case <-ctx.Done():
		slog.Info("catalog shutting down")
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			slog.Warn("catalog graceful shutdown timed out; forcing stop")
			grpcServer.Stop()
		}
	}

	slog.Info("catalog stopped cleanly")
}
