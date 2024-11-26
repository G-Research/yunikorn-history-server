package health

import (
	"context"
	"testing"

	"github.com/G-Research/unicorn-history-server/test/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HealthSuite struct {
	suite.Suite
	tp   *database.TestPostgresContainer
	pool *pgxpool.Pool
}

func (ts *HealthSuite) SetupSuite() {
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

func (ts *HealthSuite) TearDownSuite() {
	err := ts.tp.Container.Terminate(context.Background())
	require.NoError(ts.T(), err)
}

func (ts *HealthSuite) TestSubSuites() {
	ts.T().Run("HealthIntTest", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &HealthIntTest{pool: pool})
	})
	ts.T().Run("ComponentsIntTest", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &ComponentsIntTest{pool: pool})
	})

}

func TestHealthSuiteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	topSuite := new(HealthSuite)
	suite.Run(t, topSuite)
}
