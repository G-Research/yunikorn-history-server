package yunikorn

import (
	"context"
	"github.com/G-Research/yunikorn-history-server/internal/database/migrations"
	"github.com/G-Research/yunikorn-history-server/internal/database/postgres"
	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-history-server/test/database"

	"github.com/G-Research/yunikorn-history-server/test/config"
)

func TestClient_sync_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

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
		t.Fatalf("error creating postgres connection pool: %v", err)
	}
	repo, err := repository.NewPostgresRepository(pool)
	if err != nil {
		t.Fatalf("error creating postgres repository: %v", err)
	}
	eventRepository := repository.NewInMemoryEventRepository()

	c := NewRESTClient(config.GetTestYunikornConfig())
	s := NewService(ctx, repo, eventRepository, c)

	go s.Run(ctx)

	assert.Eventually(t, func() bool {
		return s.workqueue.Started()
	}, 500*time.Millisecond, 50*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	err = s.sync(ctx)
	if err != nil {
		t.Fatalf("error starting up client: %v", err)
	}

	assert.Eventually(t, func() bool {
		partitions, err := repo.GetAllPartitions(ctx)
		if err != nil {
			t.Logf("error getting partitions: %v", err)
		}
		return err == nil && len(partitions) > 0
	}, 15*time.Second, 500*time.Millisecond)

	assert.Eventually(t, func() bool {
		history, err := repo.GetApplicationsHistory(ctx)
		if err != nil {
			t.Logf("error getting applications history: %v", err)
		}
		return err == nil && len(history) > 0
	}, 15*time.Second, 500*time.Millisecond)
}
