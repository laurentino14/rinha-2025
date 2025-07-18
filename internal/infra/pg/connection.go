package pg

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laurentino14/rinha-2025/internal/config"
)

func NewPGPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, _ := pgxpool.ParseConfig(config.GetDefaultEnv("DATABASE_URL", "postgresql://rinha:rinha@postgres:5432/rinha"))
	maxConns, _ := strconv.Atoi(config.GetDefaultEnv("PG_MAX_CONNS", "25"))
	cfg.MaxConns = int32(maxConns)
	cfg.MaxConnLifetime = time.Minute * 5
	cfg.MaxConnIdleTime = time.Minute
	cfg.MinConns = 5
	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
