package repository

import (
	"context"
	"testing"
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/require"

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

	apps := []*dao.ApplicationHistoryDAOInfo{
		{
			TotalApplications: "0",
			Timestamp:         now.Add(-6 * time.Hour).UnixMilli(),
		},
		{
			TotalApplications: "2",
			Timestamp:         now.Add(-5 * time.Hour).UnixMilli(),
		},
		{
			TotalApplications: "4",
			Timestamp:         now.Add(-3 * time.Hour).UnixMilli(),
		},
		{
			TotalApplications: "7",
			Timestamp:         now.Add(-1 * time.Hour).UnixMilli(),
		},
	}

	containers := []*dao.ContainerHistoryDAOInfo{
		{
			TotalContainers: "0",
			Timestamp:       now.Add(-6 * time.Hour).UnixMilli(),
		},
		{
			TotalContainers: "2",
			Timestamp:       now.Add(-5 * time.Hour).UnixMilli(),
		},
		{
			TotalContainers: "4",
			Timestamp:       now.Add(-3 * time.Hour).UnixMilli(),
		},
		{
			TotalContainers: "7",
			Timestamp:       now.Add(-1 * time.Hour).UnixMilli(),
		},
	}

	err := repo.UpdateHistory(ctx, apps, containers)
	require.NoError(t, err)
}
