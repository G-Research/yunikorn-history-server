package repository

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/unicorn-history-server/internal/util"
)

type PartitionIntTest struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *PostgresRepository
}

func (ps *PartitionIntTest) SetupSuite() {
	ctx := context.Background()
	require.NotNil(ps.T(), ps.pool)
	repo, err := NewPostgresRepository(ps.pool)
	require.NoError(ps.T(), err)
	ps.repo = repo

	seedPartitions(ctx, ps.T(), ps.repo)
}

func (ps *PartitionIntTest) TearDownSuite() {
	ps.pool.Close()
}

func (ps *PartitionIntTest) TestGetAllPartitions() {
	ctx := context.Background()
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
		ps.Run(tt.name, func() {
			nodes, err := ps.repo.GetAllPartitions(ctx, tt.filters)
			require.NoError(ps.T(), err)
			require.Len(ps.T(), nodes, tt.expected)
		})
	}
}

func seedPartitions(ctx context.Context, t *testing.T, repo *PostgresRepository) {
	t.Helper()

	now := time.Now()
	nowNano := now.UnixNano()
	partitions := []*model.Partition{
		{
			Metadata: model.Metadata{
				CreatedAtNano: nowNano,
			},
			PartitionInfo: dao.PartitionInfo{
				ID:                      "1",
				Name:                    "default",
				ClusterID:               "cluster1",
				State:                   "Active",
				LastStateTransitionTime: now.Add(-1 * time.Hour).UnixMilli(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: nowNano,
			},
			PartitionInfo: dao.PartitionInfo{
				ID:                      "2",
				Name:                    "second",
				ClusterID:               "cluster1",
				State:                   "Active",
				LastStateTransitionTime: now.Add(-2 * time.Hour).UnixMilli(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: nowNano,
			},
			PartitionInfo: dao.PartitionInfo{
				ID:                      "3",
				Name:                    "third",
				ClusterID:               "cluster1",
				State:                   "Active",
				LastStateTransitionTime: now.Add(-3 * time.Hour).UnixMilli(),
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: nowNano,
			},
			PartitionInfo: dao.PartitionInfo{
				ID:                      "4",
				Name:                    "fourth",
				ClusterID:               "cluster1",
				State:                   "FakeState",
				LastStateTransitionTime: now.Add(-4 * time.Hour).UnixMilli(),
			},
		},
	}
	for _, p := range partitions {
		err := repo.InsertPartition(ctx, p)
		require.NoError(t, err)
	}
}
