package health

import (
	"context"
	"testing"

	"github.com/G-Research/unicorn-history-server/internal/yunikorn"
	"github.com/G-Research/unicorn-history-server/test/config"
	"github.com/G-Research/unicorn-history-server/test/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ComponentsTestSuite struct {
	suite.Suite
	tp             *database.TestPostgresContainer
	pool           *pgxpool.Pool
	yunikornClient *yunikorn.RESTClient
}

func (ts *ComponentsTestSuite) SetupSuite() {
	ctx := context.Background()
	cfg := database.InstanceConfig{
		User:     "test",
		Password: "test",
		DBName:   "template",
		Host:     "localhost",
		Port:     15436,
	}

	tp, err := database.NewTestPostgresContainer(ctx, cfg)
	require.NoError(ts.T(), err)
	ts.tp = tp
	ts.pool = tp.Pool(ctx, ts.T(), &cfg)
	ts.yunikornClient = yunikorn.NewRESTClient(config.GetTestYunikornConfig())
}

func (ts *ComponentsTestSuite) TearDownSuite() {
	err := ts.tp.Container.Terminate(context.Background())
	require.NoError(ts.T(), err)
}

func (ts *ComponentsTestSuite) TestNewComponents() {
	ctx := context.Background()

	tests := []struct {
		name               string
		component          Component
		expectedIdentifier string
		expectedHealthy    bool
	}{
		{
			name:               "should return a valid ComponentStatus when Yunikorn is reachable",
			component:          NewYunikornComponent(ts.yunikornClient),
			expectedIdentifier: "yunikorn",
			expectedHealthy:    true,
		},
		{
			name:               "should return a valid ComponentStatus when Postgres is reachable",
			component:          NewPostgresComponent(ts.pool),
			expectedIdentifier: "postgres",
			expectedHealthy:    true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			status := tt.component.Check(ctx)
			assert.Equal(ts.T(), tt.expectedIdentifier, status.Identifier)
			assert.Equal(ts.T(), tt.expectedHealthy, status.Healthy)
		})
	}
}

func TestComponentsIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	suite.Run(t, new(ComponentsTestSuite))
}
