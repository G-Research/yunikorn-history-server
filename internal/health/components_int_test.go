package health

import (
	"context"

	"github.com/G-Research/unicorn-history-server/internal/yunikorn"
	"github.com/G-Research/unicorn-history-server/test/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ComponentsIntTest struct {
	suite.Suite
	pool           *pgxpool.Pool
	yunikornClient *yunikorn.RESTClient
}

func (ts *ComponentsIntTest) SetupSuite() {
	ts.yunikornClient = yunikorn.NewRESTClient(config.GetTestYunikornConfig())
}

func (ts *ComponentsIntTest) TearDownSuite() {
	ts.pool.Close()
}

func (ts *ComponentsIntTest) TestNewComponents() {
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
