package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config хранит настройки подключения к БД
type Config struct {
	URL               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type PostgresDB struct {
	Pool *pgxpool.Pool
}

func NewPostgresDB(ctx context.Context, url string, maxConns, minConns int32, lifeTime, idleTime, checkPeriod time.Duration) (*PostgresDB, error) {
	pool, err := newPool(ctx, url, maxConns, minConns, lifeTime, idleTime, checkPeriod)
	if err != nil {
		return nil, fmt.Errorf("database.NewPostgresDB: %w", err)
	}

	return &PostgresDB{
		Pool: pool,
	}, nil
}

func newPool(ctx context.Context, url string, maxConns, minConns int32, lifeTime, idleTime, checkPeriod time.Duration) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parsing db url: %w", err)
	}

	poolCfg.MaxConns = maxConns
	poolCfg.MinConns = minConns

	poolCfg.MaxConnLifetime = lifeTime
	poolCfg.MaxConnIdleTime = idleTime

	poolCfg.HealthCheckPeriod = checkPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return pool, nil
}

func (db *PostgresDB) Ping(ctx context.Context) error {
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}

	return nil
}

func (db *PostgresDB) Close() {
	db.Pool.Close()
}
