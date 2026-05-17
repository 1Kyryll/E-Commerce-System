package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	GRPCPort    int
	HTTPPort    int
}

func Load() (Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return Config{}, errors.New("REDIS_URL is required")
	}

	grpcPort, err := parsePort("GRPC_PORT", 9000)
	if err != nil {
		return Config{}, err
	}
	httpPort, err := parsePort("HTTP_PORT", 8000)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DatabaseURL: dbURL,
		RedisURL:    redisURL,
		GRPCPort:    grpcPort,
		HTTPPort:    httpPort,
	}, nil
}

func parsePort(envKey string, def int) (int, error) {
	v := os.Getenv(envKey)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", envKey, err)
	}
	return n, nil
}
