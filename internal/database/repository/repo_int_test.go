package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/unicorn-history-server/test/database"
)

type RepositoryTestSuite struct {
	suite.Suite
	tp   *database.TestPostgresContainer
	pool *pgxpool.Pool
}

func (ts *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()
	cfg := database.InstanceConfig{
		User:     "test",
		Password: "test",
		DBName:   "template",
		Host:     "localhost",
		Port:     "15433",
	}

	tp, err := database.NewTestPostgresContainer(ctx, cfg)
	require.NoError(ts.T(), err)
	ts.tp = tp
	err = tp.Migrate("../../../migrations")
	require.NoError(ts.T(), err)

	ts.pool = tp.Pool(ctx, ts.T(), &cfg)
}

func (ts *RepositoryTestSuite) TearDownSuite() {
	err := ts.tp.Container.Terminate(context.Background())
	require.NoError(ts.T(), err)
}

func (ts *RepositoryTestSuite) TestSubSuites() {
	ts.T().Run("ApplicationTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &ApplicationTestSuite{pool: pool})
	})
	ts.T().Run("HistoryTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &HistoryTestSuite{pool: pool})
	})
	ts.T().Run("NodeTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &NodeTestSuite{pool: pool})
	})
	ts.T().Run("QueueTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &QueueTestSuite{pool: pool})
	})
	ts.T().Run("PartitionTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &PartitionTestSuite{pool: pool})
	})
}

func TestRepositoryIntegrationTestSuite(t *testing.T) {
	topSuite := new(RepositoryTestSuite)
	suite.Run(t, topSuite)
}
