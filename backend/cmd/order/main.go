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

	orderv1 "github.com/1Kyryll/ecommerce-demo/backend/gen/proto/order/v1"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/database"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/payment"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/handler"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/repo"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/service"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("order: config: %v", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, database.PoolConfig{URL: cfg.DatabaseURL})
	if err != nil {
		log.Fatalf("order: db: %v", err)
	}
	defer pool.Close()

	orderSvc := service.NewService(repo.NewRepo(pool), payment.NewFakeClient())
	orderHandler := handler.New(orderSvc)

	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(grpcServer, orderHandler)
	reflection.Register(grpcServer)

	port := cfg.GRPCPort
	if v := os.Getenv("ORDER_GRPC_PORT"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("order: ORDER_GRPC_PORT: %v", err)
		}
		port = parsed
	}
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("order: listen: %v", err)
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("order gRPC listening", "addr", addr)
		if err := grpcServer.Serve(ln); err != nil {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			log.Fatalf("order: serve: %v", err)
		}
	case <-ctx.Done():
		slog.Info("order shutting down")
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			slog.Warn("order graceful shutdown timed out; forcing stop")
			grpcServer.Stop()
		}
	}

	slog.Info("order stopped cleanly")
}
