package health

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/unicorn-history-server/internal/config"
	"github.com/G-Research/unicorn-history-server/internal/yunikorn"
	testconfig "github.com/G-Research/unicorn-history-server/test/config"
)

type HealthTestSuite struct {
	suite.Suite
	pool           *pgxpool.Pool
	yunikornClient *yunikorn.RESTClient
}

func (ts *HealthTestSuite) SetupSuite() {
	ts.yunikornClient = yunikorn.NewRESTClient(testconfig.GetTestYunikornConfig())
}

func (ts *HealthTestSuite) TearDownSuite() {
	ts.pool.Close()
}

func (ts *HealthTestSuite) TestService_Readiness() {
	ctx := context.Background()
	startedAt := time.Now()
	version := "1.0.0"

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
			startedAt:  startedAt,
			version:    version,
			components: components,
		}
		status := service.Readiness(ctx)
		expectErrorPrefix := `Get "http://invalid-host:2212/ws/v1/scheduler/healthcheck": dial tcp: lookup invalid-host`
		assert.False(ts.T(), status.Healthy)
		assert.Equal(ts.T(), 2, len(status.ComponentStatuses))
		assertStatus(ts.T(), status.ComponentStatuses, "yunikorn", false, expectErrorPrefix)
		assert.Equal(ts.T(), startedAt, status.StartedAt)
		assert.Equal(ts.T(), version, status.Version)
	})

	ts.Run("status is healthy when all components are healthy", func() {
		components := []Component{
			NewYunikornComponent(ts.yunikornClient),
			NewPostgresComponent(ts.pool),
		}
		service := Service{
			startedAt:  startedAt,
			version:    version,
			components: components,
		}
		status := service.Readiness(ctx)
		assert.True(ts.T(), status.Healthy)
		for _, componentStatus := range status.ComponentStatuses {
			assert.True(ts.T(), componentStatus.Healthy)
		}
		assert.Equal(ts.T(), startedAt, status.StartedAt)
		assert.Equal(ts.T(), version, status.Version)
	})
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
