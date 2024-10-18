package repository

import (
	"context"
	"testing"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/test/database"
)

func TestGetNodeUtilizations_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	nus := []*model.NodesUtil{
		{
			Metadata:                  model.Metadata{ID: ulid.Make().String()},
			PartitionNodesUtilDAOInfo: dao.PartitionNodesUtilDAOInfo{ClusterID: "cluster1", Partition: "default"},
		},
		{
			Metadata:                  model.Metadata{ID: ulid.Make().String()},
			PartitionNodesUtilDAOInfo: dao.PartitionNodesUtilDAOInfo{ClusterID: "cluster1", Partition: "default"},
		},
		{
			Metadata:                  model.Metadata{ID: ulid.Make().String()},
			PartitionNodesUtilDAOInfo: dao.PartitionNodesUtilDAOInfo{ClusterID: "cluster1", Partition: "default"},
		},
		{
			Metadata:                  model.Metadata{ID: ulid.Make().String()},
			PartitionNodesUtilDAOInfo: dao.PartitionNodesUtilDAOInfo{ClusterID: "cluster2", Partition: "default"},
		},
		{
			Metadata:                  model.Metadata{ID: ulid.Make().String()},
			PartitionNodesUtilDAOInfo: dao.PartitionNodesUtilDAOInfo{ClusterID: "cluster2", Partition: "default"},
		},
		{
			Metadata:                  model.Metadata{ID: ulid.Make().String()},
			PartitionNodesUtilDAOInfo: dao.PartitionNodesUtilDAOInfo{ClusterID: "cluster2", Partition: "default"},
		},
	}
	for _, nu := range nus {
		err = repo.InsertNodesUtil(ctx, nu)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		filters  NodesUtilFilters
		expected int
	}{
		{
			name: "Filter by ClusterID",
			filters: NodesUtilFilters{
				ClusterID: util.ToPtr("cluster1"),
			},
			expected: 3,
		},
		{
			name: "Filter by Partition",
			filters: NodesUtilFilters{
				Partition: util.ToPtr("default"),
			},
			expected: 6,
		},
		{
			name: "Filter By Limit",
			filters: NodesUtilFilters{
				Limit: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name: "Filter By Limit and Offset",
			filters: NodesUtilFilters{
				Limit:  util.ToPtr(10),
				Offset: util.ToPtr(3),
			},
			expected: 3,
		},
		{
			name: "Multiple filters",
			filters: NodesUtilFilters{
				ClusterID: util.ToPtr("cluster2"),
				Partition: util.ToPtr("default"),
				Limit:     util.ToPtr(1),
				Offset:    util.ToPtr(1),
			},
			expected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := repo.GetNodesUtils(ctx, tt.filters)
			require.NoError(t, err)
			require.Len(t, nodes, tt.expected)
		})
	}
}
