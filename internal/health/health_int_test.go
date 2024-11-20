package health

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/unicorn-history-server/test/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/unicorn-history-server/internal/config"
	"github.com/G-Research/unicorn-history-server/internal/yunikorn"
	testconfig "github.com/G-Research/unicorn-history-server/test/config"
)

type HealthTestSuite struct {
	suite.Suite
	tp             *database.TestPostgresContainer
	pool           *pgxpool.Pool
	yunikornClient *yunikorn.RESTClient
	startedAt      time.Time
	version        string
}

func (ts *HealthTestSuite) SetupSuite() {
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
	ts.yunikornClient = yunikorn.NewRESTClient(testconfig.GetTestYunikornConfig())

	ts.startedAt = time.Now()
	ts.version = "1.0.0"
}

func (ts *HealthTestSuite) TearDownSuite() {
	err := ts.tp.Container.Terminate(context.Background())
	require.NoError(ts.T(), err)
}

func (ts *HealthTestSuite) TestService_Readiness() {
	ctx := context.Background()

	ts.Run("status is unhealthy when one component is unhealthy", func() {
		invalidYunikornConfig := config.YunikornConfig{
			Host:   "invalid-host",
			Port:   2212,
			Secure: false,
		}
		yunikornClient := yunikorn.NewRESTClient(&invalidYunikornConfig)
		components := []Component{
			NewYunikornComponent(yunikornClient),
			NewPostgresComponent(ts.pool),
		}
		service := Service{
			startedAt:  ts.startedAt,
			version:    ts.version,
			components: components,
		}
		status := service.Readiness(ctx)
		expectErrorPrefix := `Get "http://invalid-host:2212/ws/v1/scheduler/healthcheck": dial tcp: lookup invalid-host`
		assert.False(ts.T(), status.Healthy)
		assert.Equal(ts.T(), 2, len(status.ComponentStatuses))
		assertStatus(ts.T(), status.ComponentStatuses, "yunikorn", false, expectErrorPrefix)
		assert.Equal(ts.T(), ts.startedAt, status.StartedAt)
		assert.Equal(ts.T(), ts.version, status.Version)
	})

	ts.Run("status is healthy when all components are healthy", func() {
		components := []Component{
			NewYunikornComponent(ts.yunikornClient),
			NewPostgresComponent(ts.pool),
		}
		service := Service{
			startedAt:  ts.startedAt,
			version:    ts.version,
			components: components,
		}
		status := service.Readiness(ctx)
		assert.True(ts.T(), status.Healthy)
		for _, componentStatus := range status.ComponentStatuses {
			assert.True(ts.T(), componentStatus.Healthy)
		}
		assert.Equal(ts.T(), ts.startedAt, status.StartedAt)
		assert.Equal(ts.T(), ts.version, status.Version)
	})
}

func TestHealthIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	suite.Run(t, new(HealthTestSuite))
}

func assertStatus(t *testing.T, statuses []*ComponentStatus, identifier string, expectedHealthy bool, expectedErrorPrefix string) {
	for _, status := range statuses {
		if status.Identifier == identifier {
			assert.Equal(t, expectedHealthy, status.Healthy)
			assert.Contains(t, status.Error, expectedErrorPrefix)
			return
		}
	}
	t.Fatalf("component with identifier %s not found", identifier)
}
