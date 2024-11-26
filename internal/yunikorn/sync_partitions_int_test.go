package yunikorn

import (
	"context"
	"time"

	"github.com/G-Research/unicorn-history-server/internal/database/repository"
	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SyncPartitionIntTest struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *repository.PostgresRepository
}

func (ss *SyncPartitionIntTest) SetupSuite() {
	require.NotNil(ss.T(), ss.pool)
	repo, err := repository.NewPostgresRepository(ss.pool)
	require.NoError(ss.T(), err)
	ss.repo = repo
}

func (ss *SyncPartitionIntTest) TearDownSuite() {
	ss.pool.Close()
}

func (ss *SyncPartitionIntTest) TestSyncPartitions() {
	ctx := context.Background()
	now := time.Now().UnixNano()

	tests := []struct {
		name               string
		statePartitions    []*dao.PartitionInfo
		existingPartitions []*model.Partition
		expected           []*model.Partition
		wantErr            bool
	}{
		{
			name: "Sync partition with no existing partitions in DB",
			statePartitions: []*dao.PartitionInfo{
				{
					ID:   "1",
					Name: "1",
				},
				{
					ID:   "2",
					Name: "2",
				},
			},
			existingPartitions: nil,
			expected: []*model.Partition{
				{
					PartitionInfo: dao.PartitionInfo{
						ID:   "2",
						Name: "2",
					},
				},
				{
					PartitionInfo: dao.PartitionInfo{
						ID:   "1",
						Name: "1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should mark partition 2 as deleted in DB",
			statePartitions: []*dao.PartitionInfo{
				{
					ID:   "1",
					Name: "1",
				},
			},

			existingPartitions: []*model.Partition{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:   "1",
						Name: "1",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:   "2",
						Name: "2",
					},
				},
			},
			expected: []*model.Partition{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:   "1",
						Name: "1",
					},
				},
			},
			wantErr: false,
		},
		// TODO: test syncPartition when statePartitions is nil
	}

	for _, tt := range tests {
		ss.Run(tt.name, func() {
			// clean up the table after the test
			ss.T().Cleanup(func() {
				_, err := ss.pool.Exec(ctx, "DELETE FROM partitions")
				require.NoError(ss.T(), err)
			})

			for _, partition := range tt.existingPartitions {
				err := ss.repo.InsertPartition(ctx, partition)
				require.NoError(ss.T(), err)
			}

			s := NewService(ss.repo, nil, nil)

			err := s.syncPartitions(ctx, tt.statePartitions)
			if tt.wantErr {
				require.Error(ss.T(), err)
				return
			}
			require.NoError(ss.T(), err)

			var partitionsInDB []*model.Partition
			partitionsInDB, err = s.repo.GetAllPartitions(ctx, repository.PartitionFilters{})
			require.NoError(ss.T(), err)

			for _, dbPartition := range partitionsInDB {
				found := false
				for _, expectedPartition := range tt.expected {
					if dbPartition.ID == expectedPartition.ID {
						assert.Equal(ss.T(), expectedPartition.PartitionInfo, dbPartition.PartitionInfo)
						assert.Nil(ss.T(), expectedPartition.DeletedAtNano)
						found = true
					}
				}
				if !found {
					assert.NotNil(ss.T(), dbPartition.DeletedAtNano)
				}
			}
		})
	}
}
