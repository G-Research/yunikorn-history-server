package yunikorn

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/unicorn-history-server/test/database"
)

type SyncTestSuite struct {
	suite.Suite
	tp   *database.TestPostgresContainer
	pool *pgxpool.Pool
}

func (ts *SyncTestSuite) SetupSuite() {
	ctx := context.Background()
	cfg := database.InstanceConfig{
		User:     "test",
		Password: "test",
		DBName:   "template",
		Host:     "localhost",
		Port:     15434,
	}

	tp, err := database.NewTestPostgresContainer(ctx, cfg)
	require.NoError(ts.T(), err)
	ts.tp = tp
	err = tp.Migrate("../../migrations")
	require.NoError(ts.T(), err)

	ts.pool = tp.Pool(ctx, ts.T(), &cfg)
}

func (ts *SyncTestSuite) TearDownSuite() {
	err := ts.tp.Container.Terminate(context.Background())
	require.NoError(ts.T(), err)
}

func (ts *SyncTestSuite) TestSubSuites() {
	ts.T().Run("SyncNodesTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &SyncNodesTestSuite{pool: pool})
	})
	ts.T().Run("SyncQueuesTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &SyncQueuesTestSuite{pool: pool})
	})
	ts.T().Run("SyncPartitionTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &SyncPartitionTestSuite{pool: pool})
	})
	ts.T().Run("SyncApplicationsTestSuite", func(t *testing.T) {
		pool := database.CloneDB(t, ts.tp, ts.pool)
		suite.Run(t, &SyncApplicationsTestSuite{pool: pool})
	})
}

func TestRepositoryIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	topSuite := new(SyncTestSuite)
	suite.Run(t, topSuite)
}
