package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/1Kyryll/ecommerce-demo/backend/gateway"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
)

func main() {
	// Best-effort .env load — file is optional in production.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gateway: config: %v", err)
	}

	// Default to text-format slog at debug for local dev. JSON / Info is the
	// production handler; that switch lives in a future observability plan.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	srv := gateway.NewServer(gateway.NewRouter())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("gateway starting", "addr", addr, "grpc_port", cfg.GRPCPort)
	if err := srv.ListenAndServe(ctx, addr, 5*time.Second); err != nil {
		log.Fatalf("gateway: serve: %v", err)
	}
	slog.Info("gateway stopped cleanly")
}
