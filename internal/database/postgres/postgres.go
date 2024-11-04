package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/G-Research/unicorn-history-server/internal/config"
)

func NewConnectionPool(ctx context.Context, cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, buildConnectionInfoFromConfig(cfg))
}

func BuildConnectionStringFromConfig(cfg *config.PostgresConfig) string {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)

	var additonalParams []string
	if cfg.SSLMode != "" {
		additonalParams = append(additonalParams, fmt.Sprintf("sslmode=%s", cfg.SSLMode))
	}
	if cfg.Schema != "" {
		additonalParams = append(additonalParams, fmt.Sprintf("search_path=%s", cfg.Schema))
	}

	if len(additonalParams) > 0 {
		connString += "?" + strings.Join(additonalParams, "&")
	}

	return connString
}

func buildConnectionInfoFromConfig(cfg *config.PostgresConfig) string {
	pair := func(key string, value any) string {
		return fmt.Sprintf("%s='%v'", key, value)
	}
	pairs := []string{
		pair("host", cfg.Host),
		pair("port", cfg.Port),
		pair("user", cfg.Username),
		pair("password", cfg.Password),
		pair("dbname", cfg.DbName),
	}
	if cfg.PoolMaxConns > 0 {
		pairs = append(pairs, pair("pool_max_conns", cfg.PoolMaxConns))
	}
	if cfg.PoolMinConns > 0 {
		pairs = append(pairs, pair("pool_min_conns", cfg.PoolMinConns))
	}
	if cfg.PoolMaxConnLifetime > 0 {
		pairs = append(pairs, pair("pool_max_conn_lifetime", cfg.PoolMaxConnLifetime.String()))
	}
	if cfg.PoolMaxConnIdleTime > 0 {
		pairs = append(pairs, pair("pool_max_conn_idle_time", cfg.PoolMaxConnIdleTime.String()))
	}
	if cfg.SSLMode != "" {
		pairs = append(pairs, pair("sslmode", cfg.SSLMode))
	}
	if cfg.Schema != "" {
		pairs = append(pairs, pair("search_path", cfg.Schema))
	}

	return strings.Join(pairs, " ")
}
