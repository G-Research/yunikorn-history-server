package repository

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/G-Research/yunikorn-scheduler-interface/lib/go/si"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/test/database"
)

func TestGetAllApplications_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	seedApplications(ctx, t, repo)

	tests := []struct {
		name     string
		filters  ApplicationFilters
		expected int
	}{
		{
			name: "Filter by User",
			filters: ApplicationFilters{
				User: util.ToPtr("user1"),
			},
			expected: 3,
		},
		{
			name: "Filter by Submission Time Range",
			filters: ApplicationFilters{
				SubmissionStartTime: util.ToPtr(time.Now().Add(-3 * time.Hour)),
				SubmissionEndTime:   util.ToPtr(time.Now().Add(-1 * time.Hour)),
			},
			expected: 2,
		},
		{
			name: "Filter by Finished Time Range",
			filters: ApplicationFilters{
				FinishedStartTime: util.ToPtr(time.Now().Add(-2 * time.Hour)),
				FinishedEndTime:   util.ToPtr(time.Now().Add(-1 * time.Hour)),
			},
			expected: 1,
		},
		{
			name: "Filter By Limit",
			filters: ApplicationFilters{
				Limit: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name: "Filter By Limit and Offset",
			filters: ApplicationFilters{
				Limit:  util.ToPtr(2),
				Offset: util.ToPtr(5),
			},
			expected: 1,
		},
		{
			name:     "Filter By Group",
			filters:  ApplicationFilters{Groups: []string{"group1", "group2"}},
			expected: 4,
		},
		{
			name:     "Filter By Group and User",
			filters:  ApplicationFilters{Groups: []string{"group1", "group2"}, User: util.ToPtr("user1")},
			expected: 2,
		},
		{
			name:     "No Filters",
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apps, err := repo.GetAllApplications(context.Background(), tt.filters)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, len(apps))
		})
	}
}

func TestGetAppsPerPartitionPerQueue_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	seedApplications(ctx, t, repo)

	tests := []struct {
		name     string
		filters  ApplicationFilters
		expected int
	}{
		{
			name: "Filter by User",
			filters: ApplicationFilters{
				User: util.ToPtr("user1"),
			},
			expected: 3,
		},
		{
			name: "Filter by Submission Time Range",
			filters: ApplicationFilters{
				SubmissionStartTime: util.ToPtr(time.Now().Add(-3 * time.Hour)),
				SubmissionEndTime:   util.ToPtr(time.Now().Add(-1 * time.Hour)),
			},
			expected: 2,
		},
		{
			name: "Filter by Finished Time Range",
			filters: ApplicationFilters{
				FinishedStartTime: util.ToPtr(time.Now().Add(-2 * time.Hour)),
				FinishedEndTime:   util.ToPtr(time.Now().Add(-1 * time.Hour)),
			},
			expected: 1,
		},
		{
			name: "Filter By Limit",
			filters: ApplicationFilters{
				Limit: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name: "Filter By Limit and Offset",
			filters: ApplicationFilters{
				Limit:  util.ToPtr(2),
				Offset: util.ToPtr(5),
			},
			expected: 1,
		},
		{
			name:     "Filter By Group",
			filters:  ApplicationFilters{Groups: []string{"group1", "group2"}},
			expected: 4,
		},
		{
			name:     "Filter By Group and User",
			filters:  ApplicationFilters{Groups: []string{"group1", "group2"}, User: util.ToPtr("user1")},
			expected: 2,
		},
		{
			name:     "No Filters",
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apps, err := repo.GetAppsPerPartitionPerQueue(context.Background(), "default", "root.default", tt.filters)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, len(apps))
		})
	}
}

func seedApplications(ctx context.Context, t *testing.T, repo *PostgresRepository) {
	t.Helper()

	now := time.Now()

	queues := []*dao.PartitionQueueDAOInfo{
		{
			Partition: "default",
			QueueName: "root",
			Children: []dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "root.default",
					Parent:    "root",
				},
			},
		},
	}

	err := repo.AddQueues(ctx, nil, queues)
	require.NoError(t, err, "could not seed queues")

	apps := []*model.Application{
		{
			Metadata: model.Metadata{
				ID:            "1",
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ApplicationID:   "app1",
				UsedResource:    map[string]int64{"cpu": 1},
				MaxUsedResource: map[string]int64{"cpu": 2},
				PendingResource: map[string]int64{"cpu": 1},
				Partition:       "default",
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-6 * time.Hour).UnixMilli(),
				User:            "user1",
				State:           si.EventRecord_APP_RUNNING.String(),
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "Metadata2",
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ApplicationID:   "app2",
				UsedResource:    map[string]int64{"memory": 1},
				MaxUsedResource: map[string]int64{"memory": 2},
				PendingResource: map[string]int64{"memory": 1},
				Partition:       "default",
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-5 * time.Hour).UnixMilli(),
				FinishedTime:    util.ToPtr(now.Add(-4 * time.Hour).Add(-20 * time.Minute).UnixMilli()),
				User:            "user2",
				State:           si.EventRecord_APP_COMPLETED.String(),
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "Metadata3",
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ApplicationID:   "app3",
				UsedResource:    map[string]int64{"cpu": 3},
				MaxUsedResource: map[string]int64{"cpu": 6},
				PendingResource: map[string]int64{"cpu": 3},
				Partition:       "default",
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-3 * time.Hour).UnixMilli(),
				FinishedTime:    util.ToPtr(now.Add(-1 * time.Hour).Add(-33 * time.Minute).UnixMilli()),
				User:            "user2",
				State:           si.EventRecord_APP_FAILED.String(),
				Groups:          []string{"group2"},
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "Metadata4",
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ApplicationID:   "app4",
				UsedResource:    map[string]int64{"memory": 4},
				MaxUsedResource: map[string]int64{"memory": 8},
				PendingResource: map[string]int64{"memory": 4},
				Partition:       "default",
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-2 * time.Hour).UnixMilli(),
				User:            "user3",
				State:           si.EventRecord_APP_RUNNING.String(),
				Groups:          []string{"group1"},
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "Metadata5",
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ApplicationID:   "app5",
				UsedResource:    map[string]int64{"cpu": 5},
				MaxUsedResource: map[string]int64{"cpu": 10},
				PendingResource: map[string]int64{"cpu": 5},
				Partition:       "default",
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-1 * time.Hour).UnixMilli(),
				User:            "user1",
				State:           si.EventRecord_APP_COMPLETING.String(),
				Groups:          []string{"group1", "group2"},
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "Metadata6",
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ApplicationID:   "app6",
				UsedResource:    map[string]int64{"memory": 6},
				MaxUsedResource: map[string]int64{"memory": 12},
				PendingResource: map[string]int64{"memory": 6},
				Partition:       "default",
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-43 * time.Minute).UnixMilli(),
				FinishedTime:    util.ToPtr(now.Add(-5 * time.Minute).UnixMilli()),
				User:            "user1",
				State:           si.EventRecord_APP_COMPLETED.String(),
				Groups:          []string{"group1", "group3"},
			},
		},
	}

	for _, app := range apps {
		err := repo.InsertApplication(ctx, app)
		require.NoError(t, err, "could not seed applications")
	}
}
