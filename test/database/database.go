package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	testconfig "github.com/G-Research/yunikorn-history-server/test/config"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/G-Research/yunikorn-history-server/internal/database/postgres"
	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/test/util"
)

func GetTestConnectionPool(ctx context.Context, t *testing.T, cfg *config.PostgresConfig) *pgxpool.Pool {
	pool, err := postgres.NewConnectionPool(ctx, cfg)
	if err != nil {
		t.Fatalf("could not create connection pool: %v", err)
	}
	return pool

}

func CreateTestSchema(ctx context.Context, t *testing.T) (schema string) {
	logger := log.Init(testconfig.GetTestLogConfig())

	pool := GetTestConnectionPool(ctx, t, testconfig.GetTestPostgresConfig())

	schema = fmt.Sprintf("test_%s", util.GenerateRandomAlphanum(t, 5))

	createSchemaQuery := fmt.Sprintf("CREATE SCHEMA %s", schema)
	tag, err := pool.Exec(ctx, createSchemaQuery)
	if err != nil {
		t.Fatalf("unable to create schema: %v", err)
	}
	logger.Infof("created schema %s: %s", schema, tag.String())

	setSearchPathQuery := fmt.Sprintf("SET search_path TO %s", schema)
	tag, err = pool.Exec(ctx, setSearchPathQuery)
	if err != nil {
		t.Fatalf("unable to set search path: %v", err)
	}
	logger.Infof("set search path to %s: %s", schema, tag.String())
	return
}

func DropTestSchema(ctx context.Context, t *testing.T, schema string) {
	logger := log.Init(testconfig.GetTestLogConfig())

	pool := GetTestConnectionPool(ctx, t, testconfig.GetTestPostgresConfig())

	dropQuery := fmt.Sprintf("DROP SCHEMA %s CASCADE", schema)
	tag, err := pool.Exec(ctx, dropQuery)
	if err != nil {
		t.Fatalf("unable to drop schema: %v\n", err)
	}
	logger.Infof("dropped schema %s: %s", schema, tag.String())
}
