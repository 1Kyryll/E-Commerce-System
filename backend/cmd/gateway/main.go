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
	"github.com/1Kyryll/ecommerce-demo/backend/gateway/clients"
	"github.com/1Kyryll/ecommerce-demo/backend/gateway/handlers"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/database"
	"github.com/1Kyryll/ecommerce-demo/backend/internal/observability"
)

// serviceName returns OTEL_SERVICE_NAME or the binary's default.
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
		log.Fatalf("gateway: config: %v", err)
	}

	obsShutdown, err := observability.Init(context.Background(), serviceName("gateway"))
	if err != nil {
		log.Fatalf("gateway: observability: %v", err)
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
		log.Fatalf("gateway: db: %v", err)
	}
	defer pool.Close()

	authSvc := auth.NewService(auth.NewRepo(pool), cfg.JWTSecret, cfg.JWTTTL)

	catalogAddr := os.Getenv("CATALOG_ADDR")
	if catalogAddr == "" {
		catalogAddr = "localhost:9000"
	}
	catalogClient, err := clients.DialCatalog(catalogAddr)
	if err != nil {
		log.Fatalf("gateway: catalog client: %v", err)
	}
	defer catalogClient.Close()

	cartAddr := os.Getenv("CART_ADDR")
	if cartAddr == "" {
		cartAddr = "localhost:9001"
	}
	cartClient, err := clients.DialCart(cartAddr)
	if err != nil {
		log.Fatalf("gateway: cart client: %v", err)
	}
	defer cartClient.Close()

	orderAddr := os.Getenv("ORDER_ADDR")
	if orderAddr == "" {
		orderAddr = "localhost:9002"
	}
	orderClient, err := clients.DialOrder(orderAddr)
	if err != nil {
		log.Fatalf("gateway: order client: %v", err)
	}
	defer orderClient.Close()

	productsH := handlers.NewProductHandlers(catalogClient.Client)
	cartH := handlers.NewCartHandlers(cartClient.Client)
	checkoutH := handlers.NewCheckoutHandlers(orderClient.Client, cartClient.Client)
	ordersH := handlers.NewOrderHandlers(orderClient.Client)
	secureCookies := os.Getenv("APP_ENV") == "production"

	router := gateway.NewRouter(gateway.Deps{
		AuthSvc:       authSvc,
		ProductsH:     productsH,
		CartH:         cartH,
		CheckoutH:     checkoutH,
		OrdersH:       ordersH,
		SessionTTL:    cfg.JWTTTL,
		SecureCookies: secureCookies,
	})

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	srv := gateway.NewServer(router)

	slog.Info("gateway starting", "addr", addr, "catalog_addr", catalogAddr, "cart_addr", cartAddr, "order_addr", orderAddr, "secure_cookies", secureCookies)
	if err := srv.ListenAndServe(ctx, addr, 5*time.Second); err != nil {
		log.Fatalf("gateway: serve: %v", err)
	}
	slog.Info("gateway stopped cleanly")
}
