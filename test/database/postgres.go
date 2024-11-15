package database

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type InstanceConfig struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     string
}

func (c InstanceConfig) DBConnStr() string {
	s := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.DBName)
	return s
}

type TestPostgresContainer struct {
	Container testcontainers.Container
	Cfg       InstanceConfig
	mu        sync.Mutex
}

func NewTestPostgresContainer(ctx context.Context, cfg InstanceConfig) (*TestPostgresContainer, error) {
	port := fmt.Sprintf("%s:5432/tcp", cfg.Port)
	cr := testcontainers.ContainerRequest{
		Image: "postgres:16.0-bookworm", // TODO: change to correct version
		Env: map[string]string{
			"POSTGRES_USER":     cfg.User,
			"POSTGRES_PASSWORD": cfg.Password,
			"POSTGRES_DB":       cfg.DBName,
		},
		ExposedPorts: []string{port},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: cr,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres Container: %v", err)
	}

	return &TestPostgresContainer{
		Container: container,
		Cfg:       cfg,
	}, nil
}

func (tp *TestPostgresContainer) Pool(ctx context.Context, t *testing.T, cfg *InstanceConfig) *pgxpool.Pool {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	pool, err := pgxpool.New(ctx, cfg.DBConnStr())
	require.NoError(t, err)

	return pool
}

func (tp *TestPostgresContainer) Migrate(dir string) error {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	d := "file://" + absPath
	m, err := migrate.New(d, tp.Cfg.DBConnStr())
	if err != nil {
		return fmt.Errorf("failed to create migration on conn %q: %v", tp.Cfg.DBConnStr(), err)
	}
	defer m.Close()
	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to run migration: %v", err)
	}
	return nil
}

func CloneDB(t *testing.T, tp *TestPostgresContainer, pool *pgxpool.Pool) *pgxpool.Pool {
	ctx := context.Background()
	newDBName := fmt.Sprintf("testdb_%d", time.Now().UnixNano())
	// copy the template database
	_, err := pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s;", newDBName, tp.Cfg.DBName))
	require.NoError(t, err)

	cfg := tp.Cfg
	cfg.DBName = newDBName
	pool = tp.Pool(ctx, t, &cfg)
	require.Eventually(t, func() bool {
		_, err := pool.Exec(ctx, "SELECT 1")
		return err == nil
	}, 10*time.Second, 1*time.Second)
	return pool
}
