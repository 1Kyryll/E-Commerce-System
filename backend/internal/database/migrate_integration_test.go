//go:build integration

package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// Requires: docker compose up -d postgres
// Run with: go test -tags=integration ./internal/database/...
func TestMigrateUpDown(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "pgx5://postgres:postgres@localhost:5433/ecommerce?sslmode=disable"
	}

	cwd, _ := os.Getwd()
	migrationsPath := filepath.Join(cwd, "..", "..", "migrations")
	sourceURL := "file://" + filepath.ToSlash(migrationsPath)

	if err := MigrateDown(sourceURL, dbURL); err != nil {
		t.Fatalf("initial MigrateDown: %v", err)
	}

	if err := MigrateUp(sourceURL, dbURL); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	pool, err := NewPool(context.Background(), PoolConfig{URL: "postgres://postgres:postgres@localhost:5433/ecommerce?sslmode=disable"})
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()

	wantTables := []string{"users", "products", "carts", "cart_items", "orders", "order_items", "reservations", "outbox"}
	for _, table := range wantTables {
		var exists bool
		err := pool.QueryRow(context.Background(),
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)`,
			table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("query existence of %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %q does not exist after MigrateUp", table)
		}
	}

	if err := MigrateDown(sourceURL, dbURL); err != nil {
		t.Fatalf("MigrateDown: %v", err)
	}

	for _, table := range wantTables {
		var exists bool
		err := pool.QueryRow(context.Background(),
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)`,
			table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("query existence of %s after down: %v", table, err)
		}
		if exists {
			t.Errorf("table %q still exists after MigrateDown", table)
		}
	}
}
