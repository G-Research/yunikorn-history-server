package repository

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/model"

	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/test/database"
)

func TestHistory_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	seedHistory(ctx, t, repo)

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
		t.Run(tt.name, func(t *testing.T) {
			apps, err := repo.GetApplicationsHistory(context.Background(), tt.filters)
			require.NoError(t, err)
			require.Equal(t, tt.expected, len(apps))

			containers, err := repo.GetContainersHistory(context.Background(), tt.filters)
			require.NoError(t, err)
			require.Equal(t, tt.expected, len(containers))
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
