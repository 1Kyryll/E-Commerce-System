package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
)

func main() {
	// Best-effort .env load — file is optional in production.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gateway: config: %v", err)
	}

	fmt.Printf("gateway: config loaded (grpc=:%d http=:%d)\n", cfg.GRPCPort, cfg.HTTPPort)
}
