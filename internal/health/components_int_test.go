package health

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/yunikorn-history-server/internal/database/postgres"
	"github.com/G-Research/yunikorn-history-server/internal/yunikorn"
	"github.com/G-Research/yunikorn-history-server/test/config"
)

func TestNewComponent_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	ctx := context.Background()

	yunikornClient := yunikorn.NewRESTClient(config.GetTestYunikornConfig())
	postgresPool, err := postgres.NewConnectionPool(ctx, config.GetTestPostgresConfig())
	if err != nil {
		t.Fatalf("error creating postgres connection pool: %v", err)
	}

	tests := []struct {
		name               string
		component          Component
		expectedIdentifier string
		expectedHealthy    bool
	}{
		{
			name:               "should return a valid ComponentStatus when Yunikorn is reachable",
			component:          NewYunikornComponent(yunikornClient),
			expectedIdentifier: "yunikorn",
			expectedHealthy:    true,
		},
		{
			name:               "should return a valid ComponentStatus when Postgres is reachable",
			component:          NewPostgresComponent(postgresPool),
			expectedIdentifier: "postgres",
			expectedHealthy:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.component.Check(ctx)
			assert.Equal(t, tt.expectedIdentifier, status.Identifier)
			assert.Equal(t, tt.expectedHealthy, status.Healthy)
		})
	}
}
