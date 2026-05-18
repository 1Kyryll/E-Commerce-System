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
	"github.com/1Kyryll/ecommerce-demo/backend/gateway/auth"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/database"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gateway: config: %v", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, database.PoolConfig{URL: cfg.DatabaseURL})
	if err != nil {
		log.Fatalf("gateway: db: %v", err)
	}
	defer pool.Close()

	authSvc := auth.NewService(auth.NewRepo(pool), cfg.JWTSecret, cfg.JWTTTL)
	secureCookies := os.Getenv("APP_ENV") == "production"

	router := gateway.NewRouter(gateway.Deps{
		AuthSvc:       authSvc,
		SessionTTL:    cfg.JWTTTL,
		SecureCookies: secureCookies,
	})

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	srv := gateway.NewServer(router)

	slog.Info("gateway starting", "addr", addr, "secure_cookies", secureCookies)
	if err := srv.ListenAndServe(ctx, addr, 5*time.Second); err != nil {
		log.Fatalf("gateway: serve: %v", err)
	}
	slog.Info("gateway stopped cleanly")
}
