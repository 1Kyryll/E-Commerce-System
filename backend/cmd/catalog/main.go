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
		log.Fatalf("catalog: config: %v", err)
	}

	fmt.Printf("catalog: config loaded (grpc=:%d http=:%d)\n", cfg.GRPCPort, cfg.HTTPPort)
}
