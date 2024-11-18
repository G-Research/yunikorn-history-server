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

type NodeTestSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *PostgresRepository
}

func (ns *NodeTestSuite) SetupSuite() {
	ctx := context.Background()
	require.NotNil(ns.T(), ns.pool)
	repo, err := NewPostgresRepository(ns.pool)
	require.NoError(ns.T(), err)
	ns.repo = repo

	seedNodes(ctx, ns.T(), ns.repo)
}

func (ns *NodeTestSuite) TearDownSuite() {
	ns.pool.Close()
}

func (ns *NodeTestSuite) TestGetNodesPerPartition() {
	ctx := context.Background()
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
			expected: 4,
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
				IsReserved:  util.ToPtr(true),
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
		ns.Run(tt.name, func() {
			nodes, err := ns.repo.GetNodesPerPartition(ctx, tt.partition, tt.filters)
			require.NoError(ns.T(), err)
			require.Equal(ns.T(), tt.expected, len(nodes))
		})
	}
}

func seedNodes(ctx context.Context, t *testing.T, repo *PostgresRepository) {
	t.Helper()

	now := time.Now().UnixNano()
	partition := "default"

	nodes := []*model.Node{
		{
			Metadata: model.Metadata{
				CreatedAtNano: now,
			},
			NodeDAOInfo: dao.NodeDAOInfo{
				ID:          ulid.Make().String(),
				NodeID:      "node1",
				Partition:   partition,
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
				Partition:   partition,
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
				Partition:   partition,
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
				Partition:   partition,
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
