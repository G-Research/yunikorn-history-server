package database

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/G-Research/unicorn-history-server/internal/config"
	testconfig "github.com/G-Research/unicorn-history-server/test/config"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/G-Research/unicorn-history-server/internal/database/postgres"
	"github.com/G-Research/unicorn-history-server/internal/log"
	"github.com/G-Research/unicorn-history-server/test/util"
)

// NewTestConnectionPool creates a new test schema, applies migrations and returns a connection pool to the test database.
// This function also automatically registers a cleanup function to drop the test schema after the test has run.
func NewTestConnectionPool(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()

	schema := CreateTestSchema(ctx, t)
	t.Cleanup(func() {
		DropTestSchema(ctx, t, schema)
	})

	cfg := testconfig.GetTestPostgresConfig()
	cfg.Schema = schema

	ApplyMigrations(t, cfg)

	return GetTestConnectionPool(ctx, t, cfg)
}

// ApplyMigrations applies the migrations to the test database.
func ApplyMigrations(t *testing.T, cfg *config.PostgresConfig) {
	source := "file://../../../migrations"
	connString := postgres.BuildConnectionStringFromConfig(cfg)
	m, err := migrate.New(source, connString)
	if err != nil {
		t.Fatalf("could not create migrator: %v", err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			t.Fatalf("could not apply migrations: %v", err)
		}
	}
}

// GetTestConnectionPool creates a new connection pool to the test database.
func GetTestConnectionPool(ctx context.Context, t *testing.T, cfg *config.PostgresConfig) *pgxpool.Pool {
	pool, err := postgres.NewConnectionPool(ctx, cfg)
	if err != nil {
		t.Fatalf("could not create connection pool: %v", err)
	}
	return pool
}

// CreateTestSchema creates a new test schema in the database.
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

// DropTestSchema drops the test schema from the database.
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
