package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	cartv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/cart/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/database"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/observability"
	"github.com/1Kyryll/ecommerce-demo/backend/services/cart/handler"
	"github.com/1Kyryll/ecommerce-demo/backend/services/cart/repo"
	"github.com/1Kyryll/ecommerce-demo/backend/services/cart/service"
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
		log.Fatalf("cart: config: %v", err)
	}

	obsShutdown, err := observability.Init(context.Background(), serviceName("cart"))
	if err != nil {
		log.Fatalf("cart: observability: %v", err)
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
		log.Fatalf("cart: db: %v", err)
	}
	defer pool.Close()

	cartSvc := service.NewService(repo.NewRepo(pool))
	cartHandler := handler.New(cartSvc)

	grpcServer := grpc.NewServer()
	cartv1.RegisterCartServiceServer(grpcServer, cartHandler)
	reflection.Register(grpcServer)

	port := cfg.GRPCPort
	if v := os.Getenv("CART_GRPC_PORT"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("cart: CART_GRPC_PORT: %v", err)
		}
		port = parsed
	}
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cart: listen: %v", err)
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("cart gRPC listening", "addr", addr)
		if err := grpcServer.Serve(ln); err != nil {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			log.Fatalf("cart: serve: %v", err)
		}
	case <-ctx.Done():
		slog.Info("cart shutting down")
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			slog.Warn("cart graceful shutdown timed out; forcing stop")
			grpcServer.Stop()
		}
	}

	slog.Info("cart stopped cleanly")
}
