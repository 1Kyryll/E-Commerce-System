package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/1Kyryll/ecommerce-demo/backend/internal/config"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("order: config: %v", err)
	}

	fmt.Printf("order: config loaded (grpc=:%d http=:%d)\n", cfg.GRPCPort, cfg.HTTPPort)
}
