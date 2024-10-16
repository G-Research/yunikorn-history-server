package repository

import (
	"context"
	"testing"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/test/database"
)

func TestGetNodesPerPartition_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	seedNodes(ctx, t, repo)

	tests := []struct {
		name      string
		partition string
		filters   NodeFilters
		expected  int
	}{
		{
			name:      "Filter by NodeId",
			partition: "default",
			filters: NodeFilters{
				NodeId: util.ToPtr("node1"),
			},
			expected: 1,
		},
		{
			name:      "Filter by HostName",
			partition: "default",
			filters: NodeFilters{
				HostName: util.ToPtr("host2"),
			},
			expected: 3,
		},
		{
			name:      "Filter by RackName",
			partition: "default",
			filters: NodeFilters{
				RackName: util.ToPtr("rack2"),
			},
			expected: 3,
		},
		{
			name:      "Filter by Schedulable",
			partition: "default",
			filters: NodeFilters{
				Schedulable: util.ToPtr(false),
			},
			expected: 3,
		},
		{
			name:      "Filter by IsReserved",
			partition: "default",
			filters: NodeFilters{
				IsReserved: util.ToPtr(true),
			},
			expected: 3,
		},
		{
			name:      "Filter By Limit",
			partition: "default",
			filters: NodeFilters{
				Limit: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name:      "Filter By Limit and Offset",
			partition: "default",
			filters: NodeFilters{
				Limit:  util.ToPtr(2),
				Offset: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name:      "Multiple filters",
			partition: "default",
			filters: NodeFilters{
				HostName:    util.ToPtr("host1"),
				RackName:    util.ToPtr("rack1"),
				Schedulable: util.ToPtr(true),
				IsReserved:  util.ToPtr(false),
				Limit:       util.ToPtr(4),
			},
			expected: 1,
		},
		{
			name:      "No Filters",
			partition: "default",
			expected:  4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := repo.GetNodesPerPartition(ctx, tt.partition, tt.filters)
			require.NoError(t, err)
			require.Len(t, nodes, tt.expected)
		})
	}
}

func seedNodes(ctx context.Context, t *testing.T, repo *PostgresRepository) {
	t.Helper()

	nodes := []*dao.NodeDAOInfo{
		{
			NodeID:      "node1",
			HostName:    "host1",
			RackName:    "rack1",
			Schedulable: true,
			IsReserved:  false,
		},
		{
			NodeID:      "node2",
			HostName:    "host2",
			RackName:    "rack2",
			Schedulable: false,
			IsReserved:  true,
		},
		{
			NodeID:      "node3",
			HostName:    "host2",
			RackName:    "rack2",
			Schedulable: false,
			IsReserved:  true,
		},
		{
			NodeID:      "node4",
			HostName:    "host2",
			RackName:    "rack2",
			Schedulable: false,
			IsReserved:  true,
		},
	}

	err := repo.UpsertNodes(ctx, nodes, "default")
	require.NoError(t, err)
}
