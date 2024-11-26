package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/unicorn-history-server/test/database"
)

type RepositorySuite struct {
	suite.Suite
	tp   *database.TestPostgresContainer
	pool *pgxpool.Pool
}

func (ts *RepositorySuite) SetupSuite() {
	ctx := context.Background()
	cfg := database.InstanceConfig{
		User:     "test",
		Password: "test",
		DBName:   "template",
		Host:     "localhost",
		Port:     15433,
	}

	tp, err := database.NewTestPostgresContainer(ctx, cfg)
	require.NoError(ts.T(), err)
	ts.tp = tp
	err = tp.Migrate("../../../migrations")
	require.NoError(ts.T(), err)

	ts.pool = tp.Pool(ctx, ts.T(), &cfg)
}

func (ts *RepositorySuite) TearDownSuite() {
	err := ts.tp.Container.Terminate(context.Background())
	require.NoError(ts.T(), err)
}

func (ts *RepositorySuite) TestSubSuites() {
	ts.T().Run("ApplicationIntTest", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &ApplicationIntTest{pool: pool})
	})
	ts.T().Run("HistoryIntTest", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &HistoryIntTest{pool: pool})
	})
	ts.T().Run("NodeIntTest", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &NodeIntTest{pool: pool})
	})
	ts.T().Run("QueueIntTest", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &QueueIntTest{pool: pool})
	})
	ts.T().Run("PartitionIntTest", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &PartitionIntTest{pool: pool})
	})
}

func TestRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	topSuite := new(RepositorySuite)
	suite.Run(t, topSuite)
}
