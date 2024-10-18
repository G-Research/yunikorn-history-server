package repository

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/test/database"
)

func TestGetAllPartitions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	now := time.Now()
	nowNano := now.UnixNano()
	partitions := []*model.Partition{
		{
			Metadata: model.Metadata{
				ID:            "1",
				CreatedAtNano: nowNano,
			},
			PartitionInfo: dao.PartitionInfo{
				Name:                    "default",
				ClusterID:               "cluster1",
				State:                   "Active",
				LastStateTransitionTime: now.Add(-1 * time.Hour).UnixMilli(),
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "2",
				CreatedAtNano: nowNano,
			},
			PartitionInfo: dao.PartitionInfo{
				Name:                    "second",
				ClusterID:               "cluster1",
				State:                   "Active",
				LastStateTransitionTime: now.Add(-2 * time.Hour).UnixMilli(),
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "3",
				CreatedAtNano: nowNano,
			},
			PartitionInfo: dao.PartitionInfo{
				Name:                    "third",
				ClusterID:               "cluster1",
				State:                   "Active",
				LastStateTransitionTime: now.Add(-3 * time.Hour).UnixMilli(),
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "4",
				CreatedAtNano: nowNano,
			},
			PartitionInfo: dao.PartitionInfo{
				Name:                    "fourth",
				ClusterID:               "cluster1",
				State:                   "FakeState",
				LastStateTransitionTime: now.Add(-4 * time.Hour).UnixMilli(),
			},
		},
	}
	for _, p := range partitions {
		err = repo.InsertPartition(ctx, p)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		filters  PartitionFilters
		expected int
	}{
		{
			name: "Filter by LastStateTransitionTime Time Range",
			filters: PartitionFilters{
				LastStateTransitionTimeStart: util.ToPtr(time.Now().Add(-3 * time.Hour)),
				LastStateTransitionTimeEnd:   util.ToPtr(time.Now().Add(-1 * time.Hour)),
			},
			expected: 2,
		},
		{
			name: "Filter by ClusterID",
			filters: PartitionFilters{
				ClusterID: util.ToPtr("cluster1"),
			},
			expected: 4,
		},
		{
			name: "Filter by Name",
			filters: PartitionFilters{
				Name: util.ToPtr("third"),
			},
			expected: 1,
		},
		{
			name: "Filter By Limit",
			filters: PartitionFilters{
				Limit: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name: "Filter By Limit and Offset",
			filters: PartitionFilters{
				Limit:  util.ToPtr(10),
				Offset: util.ToPtr(3),
			},
			expected: 1,
		},
		{
			name: "Filter By State",
			filters: PartitionFilters{
				State: util.ToPtr("Active"),
			},
			expected: 3,
		},
		{
			name:     "No filters",
			expected: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := repo.GetAllPartitions(ctx, tt.filters)
			require.NoError(t, err)
			require.Len(t, nodes, tt.expected)
		})
	}
}
