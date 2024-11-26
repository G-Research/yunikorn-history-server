package repository

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/G-Research/yunikorn-scheduler-interface/lib/go/si"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/unicorn-history-server/internal/util"
)

type ApplicationIntTest struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *PostgresRepository
}

func (as *ApplicationIntTest) SetupSuite() {
	ctx := context.Background()
	require.NotNil(as.T(), as.pool)
	repo, err := NewPostgresRepository(as.pool)
	require.NoError(as.T(), err)
	as.repo = repo

	seedApplications(ctx, as.T(), as.repo)
}

func (as *ApplicationIntTest) TearDownSuite() {
	as.pool.Close()
}

func (as *ApplicationIntTest) TestGetAllApplications() {
	ctx := context.Background()
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
		as.Run(tt.name, func() {
			apps, err := as.repo.GetAllApplications(ctx, tt.filters)
			require.NoError(as.T(), err)
			assert.Equal(as.T(), tt.expected, len(apps))
		})
	}
}

func (as *ApplicationIntTest) TestGetAppsPerPartitionPerQueue() {
	ctx := context.Background()
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
		as.Run(tt.name, func() {
			apps, err := as.repo.GetAppsPerPartitionPerQueue(ctx, "1", "1", tt.filters)
			require.NoError(as.T(), err)
			assert.Equal(as.T(), tt.expected, len(apps))
		})
	}
}

func seedApplications(ctx context.Context, t *testing.T, repo *PostgresRepository) {
	t.Helper()

	now := time.Now()

	partitions := []*model.Partition{
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.UnixNano(),
			},
			PartitionInfo: dao.PartitionInfo{
				ID: "1",
			},
		},
	}

	for _, p := range partitions {
		err := repo.InsertPartition(ctx, p)
		require.NoError(t, err, "could not seed partitions")
	}

	queues := []*model.Queue{
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "1",
				PartitionID: "1",
				QueueName:   "root",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "2",
				PartitionID: "1",
				Parent:      "root",
				ParentID:    util.ToPtr("1"),
				QueueName:   "root.default",
			},
		},
	}

	for _, q := range queues {
		err := repo.InsertQueue(ctx, q)
		require.NoError(t, err)
	}

	apps := []*model.Application{
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ID:              "1",
				ApplicationID:   "app1",
				UsedResource:    map[string]int64{"cpu": 1},
				MaxUsedResource: map[string]int64{"cpu": 2},
				PendingResource: map[string]int64{"cpu": 1},
				Partition:       "default",
				PartitionID:     "1",
				QueueID:         util.ToPtr("1"),
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-6 * time.Hour).UnixMilli(),
				User:            "user1",
				State:           si.EventRecord_APP_RUNNING.String(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ID:              "2",
				ApplicationID:   "app2",
				UsedResource:    map[string]int64{"memory": 1},
				MaxUsedResource: map[string]int64{"memory": 2},
				PendingResource: map[string]int64{"memory": 1},
				PartitionID:     "1",
				Partition:       "default",
				QueueID:         util.ToPtr("1"),
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-5 * time.Hour).UnixMilli(),
				FinishedTime:    util.ToPtr(now.Add(-4 * time.Hour).Add(-20 * time.Minute).UnixMilli()),
				User:            "user2",
				State:           si.EventRecord_APP_COMPLETED.String(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ID:              "3",
				ApplicationID:   "app3",
				UsedResource:    map[string]int64{"cpu": 3},
				MaxUsedResource: map[string]int64{"cpu": 6},
				PendingResource: map[string]int64{"cpu": 3},
				PartitionID:     "1",
				Partition:       "default",
				QueueID:         util.ToPtr("1"),
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
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ID:              "4",
				ApplicationID:   "app4",
				UsedResource:    map[string]int64{"memory": 4},
				MaxUsedResource: map[string]int64{"memory": 8},
				PendingResource: map[string]int64{"memory": 4},
				PartitionID:     "1",
				Partition:       "default",
				QueueID:         util.ToPtr("1"),
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-2 * time.Hour).UnixMilli(),
				User:            "user3",
				State:           si.EventRecord_APP_RUNNING.String(),
				Groups:          []string{"group1"},
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ID:              "5",
				ApplicationID:   "app5",
				UsedResource:    map[string]int64{"cpu": 5},
				MaxUsedResource: map[string]int64{"cpu": 10},
				PendingResource: map[string]int64{"cpu": 5},
				PartitionID:     "1",
				Partition:       "default",
				QueueID:         util.ToPtr("1"),
				QueueName:       "root.default",
				SubmissionTime:  now.Add(-1 * time.Hour).UnixMilli(),
				User:            "user1",
				State:           si.EventRecord_APP_COMPLETING.String(),
				Groups:          []string{"group1", "group2"},
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now.UnixNano(),
			},
			ApplicationDAOInfo: dao.ApplicationDAOInfo{
				ID:              "6",
				ApplicationID:   "app6",
				UsedResource:    map[string]int64{"memory": 6},
				MaxUsedResource: map[string]int64{"memory": 12},
				PendingResource: map[string]int64{"memory": 6},
				PartitionID:     "1",
				Partition:       "default",
				QueueID:         util.ToPtr("1"),
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
