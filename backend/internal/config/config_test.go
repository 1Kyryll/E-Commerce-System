package config

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-at-least-32-chars-1234"

func TestLoadFromEnv_Success(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("GRPC_PORT", "9001")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("JWT_SECRET", testSecret)
	t.Setenv("JWT_TTL", "12h")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if c.DatabaseURL != "postgres://u:p@localhost:5432/db" {
		t.Errorf("DatabaseURL = %q", c.DatabaseURL)
	}
	if c.RedisURL != "redis://localhost:6379" {
		t.Errorf("RedisURL = %q", c.RedisURL)
	}
	if c.GRPCPort != 9001 {
		t.Errorf("GRPCPort = %d", c.GRPCPort)
	}
	if c.HTTPPort != 8080 {
		t.Errorf("HTTPPort = %d", c.HTTPPort)
	}
	if string(c.JWTSecret) != testSecret {
		t.Errorf("JWTSecret = %q", string(c.JWTSecret))
	}
	if c.JWTTTL != 12*time.Hour {
		t.Errorf("JWTTTL = %v, want 12h", c.JWTTTL)
	}
}

func TestLoadFromEnv_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_SECRET", testSecret)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL, got nil")
	}
}

func TestLoadFromEnv_DefaultPortsAndTTL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("JWT_SECRET", testSecret)
	t.Setenv("JWT_TTL", "")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if c.GRPCPort != 9000 {
		t.Errorf("default GRPCPort = %d, want 9000", c.GRPCPort)
	}
	if c.HTTPPort != 8000 {
		t.Errorf("default HTTPPort = %d, want 8000", c.HTTPPort)
	}
	if c.JWTTTL != 24*time.Hour {
		t.Errorf("default JWTTTL = %v, want 24h", c.JWTTTL)
	}
}

func TestLoadFromEnv_MissingJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET, got nil")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("error %q does not mention JWT_SECRET", err.Error())
	}
}

func TestLoadFromEnv_ShortJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_SECRET", "short")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for short JWT_SECRET")
	}
}

func TestLoadFromEnv_BadJWTTTL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_SECRET", testSecret)
	t.Setenv("JWT_TTL", "not-a-duration")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for malformed JWT_TTL")
	}
}
