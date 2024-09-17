package migrations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/yunikorn-history-server/test/config"
	"github.com/G-Research/yunikorn-history-server/test/database"
)

func TestGoMigrate_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	schema := database.CreateTestSchema(ctx, t)
	t.Cleanup(func() {
		database.DropTestSchema(ctx, t, schema)
	})

	cfg := config.GetTestPostgresConfig()
	cfg.Schema = schema
	m, err := New(cfg, "../../../migrations")
	if err != nil {
		t.Fatalf("could not create migrator: %v", err)
	}

	applied, err := m.Up()
	assert.Truef(t, applied, "expected up migrations to be applied for the first run")
	if err != nil {
		t.Fatalf("error running migrations up: %v", err)
	}

	applied, err = m.Up()
	assert.Falsef(t, applied, "expected no up migrations to be applied for the second run")
	if err != nil {
		t.Fatalf("error running migrations up: %v", err)
	}

	applied, err = m.Down()
	assert.Truef(t, applied, "expected down migrations to be applied for the first run")
	if err != nil {
		t.Fatalf("error running migrations down: %v", err)
	}

	applied, err = m.Down()
	assert.Falsef(t, applied, "expected no down migrations to be applied for the second run")
	if err != nil {
		t.Fatalf("error running migrations down: %v", err)
	}
}
