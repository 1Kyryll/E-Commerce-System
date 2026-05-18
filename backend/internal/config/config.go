package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

const minJWTSecretLen = 32

type Config struct {
	DatabaseURL string
	RedisURL    string
	GRPCPort    int
	HTTPPort    int
	JWTSecret   []byte
	JWTTTL      time.Duration
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

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if len(secret) < minJWTSecretLen {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least %d characters", minJWTSecretLen)
	}

	ttl := 24 * time.Hour
	if v := os.Getenv("JWT_TTL"); v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("JWT_TTL: %w", err)
		}
		ttl = parsed
	}

	return Config{
		DatabaseURL: dbURL,
		RedisURL:    redisURL,
		GRPCPort:    grpcPort,
		HTTPPort:    httpPort,
		JWTSecret:   []byte(secret),
		JWTTTL:      ttl,
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
