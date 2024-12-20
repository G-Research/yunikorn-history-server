package repository

import (
	"context"
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
	require.NotNil(ps.T(), ps.pool)
	repo, err := NewPostgresRepository(ps.pool)
	require.NoError(ps.T(), err)
	ps.repo = repo
}

func (ps *PartitionIntTest) TearDownSuite() {
	ps.pool.Close()
}

func (ps *PartitionIntTest) TestInsertPartition() {
	ctx := context.Background()
	now := time.Now()
	nowNano := now.UnixNano()
	tests := []struct {
		name               string
		existingPartitions []*model.Partition
		partitionToInsert  *model.Partition
		expectedError      bool
	}{
		{
			name:               "Insert Partition",
			existingPartitions: nil,
			partitionToInsert: &model.Partition{
				Metadata: model.Metadata{
					CreatedAtNano: nowNano,
				},
				PartitionInfo: dao.PartitionInfo{
					ID:                      "300",
					Name:                    "default",
					ClusterID:               "cluster1",
					State:                   "Active",
					LastStateTransitionTime: now.UnixMilli(),
				},
			},
			expectedError: false,
		},
		{
			name: "Insert Partition with same ID",
			existingPartitions: []*model.Partition{
				{
					Metadata: model.Metadata{
						CreatedAtNano: nowNano,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:                      "300",
						Name:                    "default",
						ClusterID:               "cluster1",
						State:                   "Active",
						LastStateTransitionTime: now.UnixMilli(),
					},
				},
			},
			partitionToInsert: &model.Partition{
				Metadata: model.Metadata{
					CreatedAtNano: nowNano,
				},
				PartitionInfo: dao.PartitionInfo{
					ID:                      "300",
					Name:                    "default",
					ClusterID:               "cluster1",
					State:                   "Active",
					LastStateTransitionTime: now.UnixMilli(),
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		ps.Run(tt.name, func() {
			if tt.existingPartitions != nil {
				for _, p := range tt.existingPartitions {
					err := ps.repo.InsertPartition(ctx, p)
					require.NoError(ps.T(), err)
				}
			}
			err := ps.repo.InsertPartition(ctx, tt.partitionToInsert)
			require.Equal(ps.T(), tt.expectedError, err != nil)
			ps.clearPartitionsTable(ctx)
		})
	}
}

func (ps *PartitionIntTest) TestGetAllPartitions() {
	ctx := context.Background()
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
		err := ps.repo.InsertPartition(ctx, p)
		require.NoError(ps.T(), err)
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
		ps.Run(tt.name, func() {
			nodes, err := ps.repo.GetAllPartitions(ctx, tt.filters)
			require.NoError(ps.T(), err)
			require.Len(ps.T(), nodes, tt.expected)
		})
	}
	ps.clearPartitionsTable(ctx)
}

func (ps *PartitionIntTest) TestGetPartitionByID() {
	ctx := context.Background()
	now := time.Now()
	nowNano := now.UnixNano()

	tests := []struct {
		name              string
		existingPartition []*model.Partition
		partitionID       string
		expectedError     bool
	}{
		{
			name: "Get Partition by ID",
			existingPartition: []*model.Partition{
				{
					Metadata: model.Metadata{
						CreatedAtNano: nowNano,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:                      "100",
						Name:                    "default",
						ClusterID:               "cluster1",
						State:                   "Active",
						LastStateTransitionTime: now.Add(-1 * time.Hour).UnixMilli(),
					},
				},
			},
			partitionID:   "100",
			expectedError: false,
		},
		{
			name:              "Get Partition by ID when partition does not exist",
			existingPartition: nil,
			partitionID:       "200",
			expectedError:     true,
		},
	}

	for _, tt := range tests {
		ps.Run(tt.name, func() {
			if tt.existingPartition != nil {
				for _, p := range tt.existingPartition {
					err := ps.repo.InsertPartition(ctx, p)
					require.NoError(ps.T(), err)
				}
			}
			partition, err := ps.repo.GetPartitionByID(ctx, tt.partitionID)
			require.Equal(ps.T(), tt.expectedError, err != nil)
			if !tt.expectedError {
				require.Equal(ps.T(), tt.partitionID, partition.ID)
			}
			ps.clearPartitionsTable(ctx)
		})
	}
}

func (ps *PartitionIntTest) TestUpdatePartition() {
	ctx := context.Background()
	nowNano := time.Now().UnixNano()
	tests := []struct {
		name              string
		existingPartition []*model.Partition
		partitionToUpdate *model.Partition
		expectedError     bool
	}{
		{
			name: "Update Partition when exists",
			existingPartition: []*model.Partition{
				{
					Metadata: model.Metadata{
						CreatedAtNano: nowNano,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:                      "100",
						Name:                    "default",
						ClusterID:               "cluster1",
						State:                   "Active",
						LastStateTransitionTime: time.Now().Add(-1 * time.Hour).UnixMilli(),
					},
				},
			},
			partitionToUpdate: &model.Partition{
				Metadata: model.Metadata{
					CreatedAtNano: nowNano,
				},
				PartitionInfo: dao.PartitionInfo{
					ID:                      "100",
					Name:                    "updated",
					ClusterID:               "cluster2",
					State:                   "Inactive",
					LastStateTransitionTime: time.Now().UnixMilli(),
				},
			},
			expectedError: false,
		},
		{
			name:              "Update Partition when does not exist",
			existingPartition: nil,
			partitionToUpdate: &model.Partition{
				Metadata: model.Metadata{
					CreatedAtNano: nowNano,
				},
				PartitionInfo: dao.PartitionInfo{
					ID:                      "200",
					Name:                    "updated",
					ClusterID:               "cluster2",
					State:                   "Inactive",
					LastStateTransitionTime: time.Now().UnixMilli(),
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		ps.Run(tt.name, func() {
			if tt.existingPartition != nil {
				for _, p := range tt.existingPartition {
					err := ps.repo.InsertPartition(ctx, p)
					require.NoError(ps.T(), err)
				}
			}

			err := ps.repo.UpdatePartition(ctx, tt.partitionToUpdate)
			require.Equal(ps.T(), tt.expectedError, err != nil)

			if !tt.expectedError {
				partition, err := ps.repo.GetPartitionByID(ctx, tt.partitionToUpdate.ID)
				require.NoError(ps.T(), err)
				require.Equal(ps.T(), tt.partitionToUpdate.Name, partition.Name)
				require.Equal(ps.T(), tt.partitionToUpdate.ClusterID, partition.ClusterID)
				require.Equal(ps.T(), tt.partitionToUpdate.State, partition.State)
				require.Equal(ps.T(), tt.partitionToUpdate.LastStateTransitionTime, partition.LastStateTransitionTime)
			}
			ps.clearPartitionsTable(ctx)
		})
	}
}

func (ps *PartitionIntTest) TestDeletePartitionsNotInIDs() {
	ctx := context.Background()
	now := time.Now()
	nowNano := now.UnixNano()

	tests := []struct {
		name               string
		existingPartitions []*model.Partition
		partitionIDs       []string
		expectedError      bool
	}{
		{
			name: "Delete Partitions with correct IDs",
			existingPartitions: []*model.Partition{
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
			},
			partitionIDs:  []string{"1", "2"},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		ps.Run(tt.name, func() {
			for _, p := range tt.existingPartitions {
				err := ps.repo.InsertPartition(ctx, p)
				require.NoError(ps.T(), err)
			}

			err := ps.repo.DeletePartitionsNotInIDs(ctx, tt.partitionIDs, nowNano)
			require.Equal(ps.T(), tt.expectedError, err != nil)
			ps.clearPartitionsTable(ctx)
		})
	}
}

func (ps *PartitionIntTest) clearPartitionsTable(ctx context.Context) {
	_, err := ps.pool.Exec(ctx, "DELETE FROM partitions")
	require.NoError(ps.T(), err)
}
