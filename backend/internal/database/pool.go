package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultMaxConns is the per-service pool size when neither PoolConfig.MaxConns
// nor DATABASE_MAX_CONNS env var is set. pgx's own default is 4, which is too
// small for any realistic load; 25 gives 4 services × 25 = 100 simultaneous
// connections in steady state (Postgres's default max_connections is 100;
// docker-compose raises it to 200 to leave headroom).
const DefaultMaxConns int32 = 25

type PoolConfig struct {
	URL      string
	MaxConns int32
}

func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	max := cfg.MaxConns
	if max <= 0 {
		if v := os.Getenv("DATABASE_MAX_CONNS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				max = int32(n)
			}
		} else {
			max = DefaultMaxConns
		}
	}
	pcfg.MaxConns = max

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
