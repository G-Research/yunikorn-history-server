package yunikorn

import (
	"context"
	"github.com/G-Research/yunikorn-history-server/internal/database/migrations"
	"github.com/G-Research/yunikorn-history-server/internal/database/postgres"
	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"testing"

	"github.com/G-Research/yunikorn-history-server/test/database"

	"github.com/G-Research/yunikorn-history-server/test/config"
)

func TestClient_sync_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()

	schema := database.CreateTestSchema(ctx, t)
	t.Cleanup(func() {
		database.DropTestSchema(ctx, t, schema)
	})

	cfg := config.GetTestPostgresConfig()
	cfg.Schema = schema
	m, err := migrations.New(cfg, "../../migrations")
	if err != nil {
		t.Fatalf("error creating migrator: %v", err)
	}
	applied, err := m.Up()
	if err != nil {
		t.Fatalf("error occured while applying migrations: %v", err)
	}
	if !applied {
		t.Fatal("migrator finished but migrations were not applied")
	}

	pool, err := postgres.NewConnectionPool(ctx, cfg)
	if err != nil {
		t.Errorf("error creating postgres connection pool: %v", err)
	}
	repo, err := repository.NewPostgresRepository(pool)
	if err != nil {
		t.Errorf("error creating postgres repository: %v", err)
	}
	eventRepository := repository.NewInMemoryEventRepository()

	c := NewRESTClient(config.GetTestYunikornConfig())
	s := NewService(repo, eventRepository, c)

	err = s.sync(ctx)
	if err != nil {
		t.Errorf("error starting up client: %v", err)
	}
}
