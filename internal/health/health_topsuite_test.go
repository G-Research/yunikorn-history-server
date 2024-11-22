package health

import (
	"context"
	"testing"

	"github.com/G-Research/unicorn-history-server/test/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HealthToptSuite struct {
	suite.Suite
	tp   *database.TestPostgresContainer
	pool *pgxpool.Pool
}

func (ts *HealthToptSuite) SetupSuite() {
	ctx := context.Background()
	cfg := database.InstanceConfig{
		User:     "test",
		Password: "test",
		DBName:   "template",
		Host:     "localhost",
		Port:     15437,
	}

	tp, err := database.NewTestPostgresContainer(ctx, cfg)
	require.NoError(ts.T(), err)
	ts.tp = tp
	ts.pool = tp.Pool(ctx, ts.T(), &cfg)
}

func (ts *HealthToptSuite) TearDownSuite() {
	err := ts.tp.Container.Terminate(context.Background())
	require.NoError(ts.T(), err)
}

func (ts *HealthToptSuite) TestSubSuites() {
	ts.T().Run("HealthTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &HealthTestSuite{pool: pool})
	})
	ts.T().Run("ComponentsTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &ComponentsTestSuite{pool: pool})
	})

}

func TestHealthIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	topSuite := new(HealthToptSuite)
	suite.Run(t, topSuite)
}
