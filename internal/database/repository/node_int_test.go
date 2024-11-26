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

type NodeIntTest struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *PostgresRepository
}

var partitionID = ulid.Make().String()

func (ns *NodeIntTest) SetupSuite() {
	ctx := context.Background()
	require.NotNil(ns.T(), ns.pool)
	repo, err := NewPostgresRepository(ns.pool)
	require.NoError(ns.T(), err)
	ns.repo = repo

	seedNodes(ctx, ns.T(), ns.repo)
}

func (ns *NodeIntTest) TearDownSuite() {
	ns.pool.Close()
}

func (ns *NodeIntTest) TestGetNodesPerPartition() {
	ctx := context.Background()
	tests := []struct {
		name        string
		partitionID string
		filters     NodeFilters
		expected    int
	}{
		{
			name:        "Filter by NodeId",
			partitionID: partitionID,
			filters: NodeFilters{
				NodeId: util.ToPtr("node1"),
			},
			expected: 1,
		},
		{
			name:        "Filter by HostName",
			partitionID: partitionID,
			filters: NodeFilters{
				HostName: util.ToPtr("host2"),
			},
			expected: 3,
		},
		{
			name:        "Filter by RackName",
			partitionID: partitionID,
			filters: NodeFilters{
				RackName: util.ToPtr("rack2"),
			},
			expected: 3,
		},
		{
			name:        "Filter by Schedulable",
			partitionID: partitionID,
			filters: NodeFilters{
				Schedulable: util.ToPtr(false),
			},
			expected: 3,
		},
		{
			name:        "Filter by IsReserved",
			partitionID: partitionID,
			filters: NodeFilters{
				IsReserved: util.ToPtr(true),
			},
			expected: 4,
		},
		{
			name:        "Filter By Limit",
			partitionID: partitionID,
			filters: NodeFilters{
				Limit: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name:        "Filter By Limit and Offset",
			partitionID: partitionID,
			filters: NodeFilters{
				Limit:  util.ToPtr(2),
				Offset: util.ToPtr(2),
			},
			expected: 2,
		},
		{
			name:        "Multiple filters",
			partitionID: partitionID,
			filters: NodeFilters{
				HostName:    util.ToPtr("host1"),
				RackName:    util.ToPtr("rack1"),
				Schedulable: util.ToPtr(true),
				IsReserved:  util.ToPtr(true),
				Limit:       util.ToPtr(4),
			},
			expected: 1,
		},
		{
			name:        "No Filters",
			partitionID: partitionID,
			expected:    4,
		},
	}

	for _, tt := range tests {
		ns.Run(tt.name, func() {
			nodes, err := ns.repo.GetNodesPerPartition(ctx, tt.partitionID, tt.filters)
			require.NoError(ns.T(), err)
			require.Equal(ns.T(), tt.expected, len(nodes))
		})
	}
}

func seedNodes(ctx context.Context, t *testing.T, repo *PostgresRepository) {
	t.Helper()

	now := time.Now().UnixNano()

	nodes := []*model.Node{
		{
			Metadata: model.Metadata{
				CreatedAtNano: now,
			},
			NodeDAOInfo: dao.NodeDAOInfo{
				ID:          ulid.Make().String(),
				NodeID:      "node1",
				PartitionID: partitionID,
				HostName:    "host1",
				RackName:    "rack1",
				Schedulable: true,
				IsReserved:  true,
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now,
			},
			NodeDAOInfo: dao.NodeDAOInfo{
				ID:          ulid.Make().String(),
				NodeID:      "node2",
				PartitionID: partitionID,
				HostName:    "host2",
				RackName:    "rack2",
				Schedulable: false,
				IsReserved:  true,
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now,
			},
			NodeDAOInfo: dao.NodeDAOInfo{
				ID:          ulid.Make().String(),
				NodeID:      "node3",
				PartitionID: partitionID,
				HostName:    "host2",
				RackName:    "rack2",
				Schedulable: false,
				IsReserved:  true,
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: now,
			},
			NodeDAOInfo: dao.NodeDAOInfo{
				NodeID:      "node4",
				ID:          ulid.Make().String(),
				HostName:    "host2",
				PartitionID: partitionID,
				RackName:    "rack2",
				Schedulable: false,
				IsReserved:  true,
			},
		},
	}

	for _, node := range nodes {
		err := repo.InsertNode(ctx, node)
		require.NoError(t, err)
	}
}
