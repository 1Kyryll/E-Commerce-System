package config

import (
	"testing"
)

func TestLoadFromEnv_Success(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("GRPC_PORT", "9001")
	t.Setenv("HTTP_PORT", "8080")

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
}

func TestLoadFromEnv_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "redis://localhost:6379")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL, got nil")
	}
}

func TestLoadFromEnv_DefaultPorts(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("HTTP_PORT", "")

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
}
