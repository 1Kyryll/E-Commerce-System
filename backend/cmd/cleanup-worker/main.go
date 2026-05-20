package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/database"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/repo"
	"github.com/1Kyryll/ecommerce-demo/backend/services/order/service"
)

const (
	cleanupInterval = 60 * time.Second
	cleanupBatch    = 100
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("cleanup-worker: config: %v", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, database.PoolConfig{URL: cfg.DatabaseURL})
	if err != nil {
		log.Fatalf("cleanup-worker: db: %v", err)
	}
	defer pool.Close()

	svc := service.NewService(repo.NewRepo(pool))

	slog.Info("cleanup-worker starting", "interval", cleanupInterval, "batch", cleanupBatch)

	// Run once immediately, then on each tick.
	if err := runOnce(ctx, svc); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("cleanup tick failed", "err", err)
	}

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("cleanup-worker stopped cleanly")
			return
		case <-ticker.C:
			if err := runOnce(ctx, svc); err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("cleanup tick failed", "err", err)
			}
		}
	}
}

func runOnce(ctx context.Context, svc *service.Service) error {
	n, err := svc.CleanupExpired(ctx, cleanupBatch)
	if err != nil {
		return err
	}
	if n > 0 {
		slog.Info("released expired reservations", "count", n)
	}
	return nil
}
