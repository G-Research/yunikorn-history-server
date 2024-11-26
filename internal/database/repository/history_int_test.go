package repository

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/unicorn-history-server/internal/util"
)

type HistoryIntTest struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *PostgresRepository
}

func (hs *HistoryIntTest) SetupSuite() {
	ctx := context.Background()
	require.NotNil(hs.T(), hs.pool)
	repo, err := NewPostgresRepository(hs.pool)
	require.NoError(hs.T(), err)
	hs.repo = repo

	seedHistory(ctx, hs.T(), hs.repo)
}

func (hs *HistoryIntTest) TearDownSuite() {
	hs.pool.Close()
}

func (hs *HistoryIntTest) TestGetApplicationsHistory() {
	ctx := context.Background()
	tests := []struct {
		name     string
		filters  HistoryFilters
		expected int
	}{
		{
			name: "Filter by Timestamp Range",
			filters: HistoryFilters{
				TimestampStart: util.ToPtr(time.Now().Add(-8 * time.Hour)),
				TimestampEnd:   util.ToPtr(time.Now().Add(-2 * time.Hour)),
			},
			expected: 3,
		},
		{
			name: "Filter by Limit",
			filters: HistoryFilters{
				Limit: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name: "Filter by Offset",
			filters: HistoryFilters{
				Limit:  util.ToPtr(2),
				Offset: util.ToPtr(3),
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		hs.Run(tt.name, func() {
			apps, err := hs.repo.GetApplicationsHistory(ctx, tt.filters)
			require.NoError(hs.T(), err)
			require.Equal(hs.T(), tt.expected, len(apps))
		})
	}
}

func (hs *HistoryIntTest) TestGetContainersHistory() {
	ctx := context.Background()
	tests := []struct {
		name     string
		filters  HistoryFilters
		expected int
	}{
		{
			name: "Filter by Timestamp Range",
			filters: HistoryFilters{
				TimestampStart: util.ToPtr(time.Now().Add(-8 * time.Hour)),
				TimestampEnd:   util.ToPtr(time.Now().Add(-2 * time.Hour)),
			},
			expected: 3,
		},
		{
			name: "Filter by Limit",
			filters: HistoryFilters{
				Limit: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name: "Filter by Offset",
			filters: HistoryFilters{
				Limit:  util.ToPtr(2),
				Offset: util.ToPtr(3),
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		hs.Run(tt.name, func() {
			containers, err := hs.repo.GetContainersHistory(ctx, tt.filters)
			require.NoError(hs.T(), err)
			require.Equal(hs.T(), tt.expected, len(containers))
		})
	}
}

func seedHistory(ctx context.Context, t *testing.T, repo *PostgresRepository) {
	t.Helper()

	now := time.Now()

	apps := []*model.AppHistory{
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.Add(-6 * time.Hour).UnixNano(),
			},
			ID: ulid.Make().String(),
			ApplicationHistoryDAOInfo: dao.ApplicationHistoryDAOInfo{
				TotalApplications: "0",
				Timestamp:         now.Add(-6 * time.Hour).UnixNano(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.Add(-5 * time.Hour).UnixNano(),
			},
			ID: ulid.Make().String(),
			ApplicationHistoryDAOInfo: dao.ApplicationHistoryDAOInfo{
				TotalApplications: "2",
				Timestamp:         now.Add(-5 * time.Hour).UnixNano(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.Add(-3 * time.Hour).UnixNano(),
			},
			ID: ulid.Make().String(),
			ApplicationHistoryDAOInfo: dao.ApplicationHistoryDAOInfo{
				TotalApplications: "5",
				Timestamp:         now.Add(-3 * time.Hour).UnixNano(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.Add(-1 * time.Hour).UnixNano(),
			},
			ID: ulid.Make().String(),
			ApplicationHistoryDAOInfo: dao.ApplicationHistoryDAOInfo{
				TotalApplications: "7",
				Timestamp:         now.Add(-1 * time.Hour).UnixNano(),
			},
		},
	}

	containers := []*model.ContainerHistory{
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.Add(-6 * time.Hour).UnixNano(),
			},
			ID: ulid.Make().String(),
			ContainerHistoryDAOInfo: dao.ContainerHistoryDAOInfo{
				TotalContainers: "0",
				Timestamp:       now.Add(-6 * time.Hour).UnixNano(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.Add(-5 * time.Hour).UnixNano(),
			},
			ID: ulid.Make().String(),
			ContainerHistoryDAOInfo: dao.ContainerHistoryDAOInfo{
				TotalContainers: "2",
				Timestamp:       now.Add(-5 * time.Hour).UnixNano(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.Add(-3 * time.Hour).UnixNano(),
			},
			ID: ulid.Make().String(),
			ContainerHistoryDAOInfo: dao.ContainerHistoryDAOInfo{
				TotalContainers: "5",
				Timestamp:       now.Add(-3 * time.Hour).UnixNano(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.Add(-1 * time.Hour).UnixNano(),
			},
			ID: ulid.Make().String(),
			ContainerHistoryDAOInfo: dao.ContainerHistoryDAOInfo{
				TotalContainers: "7",
				Timestamp:       now.Add(-1 * time.Hour).UnixNano(),
			},
		},
	}

	for _, app := range apps {
		err := repo.InsertAppHistory(ctx, app)
		require.NoError(t, err)
	}
	for _, container := range containers {
		err := repo.InsertContainerHistory(ctx, container)
		require.NoError(t, err)
	}
}
