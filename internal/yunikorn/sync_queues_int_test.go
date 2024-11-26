package yunikorn

import (
	"context"
	"time"

	"github.com/G-Research/unicorn-history-server/internal/database/repository"
	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SyncQueuesIntTest struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *repository.PostgresRepository
}

func (ss *SyncQueuesIntTest) SetupSuite() {
	require.NotNil(ss.T(), ss.pool)
	repo, err := repository.NewPostgresRepository(ss.pool)
	require.NoError(ss.T(), err)
	ss.repo = repo
}

func (ss *SyncQueuesIntTest) TearDownSuite() {
	ss.pool.Close()
}

func (ss *SyncQueuesIntTest) TestSyncQueues() {

	ctx := context.Background()
	now := time.Now().UnixNano()

	tests := []struct {
		name           string
		stateQueues    []dao.PartitionQueueDAOInfo
		existingQueues []*model.Queue
		expected       []*model.Queue
		wantErr        bool
	}{
		{
			name: "Sync queues with no existing queues",
			stateQueues: []dao.PartitionQueueDAOInfo{
				{
					ID:          "1",
					QueueName:   "root",
					PartitionID: "1",
					Children: []dao.PartitionQueueDAOInfo{
						{
							ID:          "2",
							QueueName:   "root.child-1",
							PartitionID: "1",
							Children: []dao.PartitionQueueDAOInfo{
								{
									ID:          "3",
									QueueName:   "root.child-1.1",
									PartitionID: "1",
								},
								{
									ID:          "4",
									QueueName:   "root.child-1.2",
									PartitionID: "1",
								},
							},
						},
					},
				},
			},
			existingQueues: nil,
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "1",
						QueueName:   "root",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "2",
						QueueName:   "root.child-1",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "3",
						QueueName:   "root.child-1.1",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "4",
						QueueName:   "root.child-1.2",
						PartitionID: "1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Sync queues with existing queues in DB",
			stateQueues: []dao.PartitionQueueDAOInfo{{
				ID:          "1",
				QueueName:   "root",
				PartitionID: "1",
				Children: []dao.PartitionQueueDAOInfo{
					{
						ID:          "2",
						QueueName:   "root.child-1",
						PartitionID: "1",
					},
					{
						ID:          "3",
						QueueName:   "root.child-2",
						PartitionID: "1",
					},
				},
			},
			},

			existingQueues: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "1",
						QueueName:   "root",
						PartitionID: "1",
					},
				},
			},
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "1",
						QueueName:   "root",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "2",
						QueueName:   "root.child-1",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "3",
						QueueName:   "root.child-2",
						PartitionID: "1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Sync queues when queue is deleted",
			stateQueues: []dao.PartitionQueueDAOInfo{
				{
					ID:          "1",
					QueueName:   "root",
					PartitionID: "1",
					Children: []dao.PartitionQueueDAOInfo{
						{
							ID:          "3",
							QueueName:   "root.child-2",
							PartitionID: "1",
						},
					},
				},
			},
			existingQueues: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "1",
						QueueName:   "root",
						PartitionID: "1",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "2",
						QueueName:   "root.child-1",
						PartitionID: "1",
					},
				},
			},
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "1",
						QueueName:   "root",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "3",
						QueueName:   "root.child-2",
						PartitionID: "1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Sync queues with multiple partitions",
			stateQueues: []dao.PartitionQueueDAOInfo{
				{
					ID:          "1",
					QueueName:   "root",
					PartitionID: "1",
					Children: []dao.PartitionQueueDAOInfo{
						{
							ID:          "2",
							QueueName:   "root.child-1",
							PartitionID: "1",
						},
						{
							ID:          "3",
							QueueName:   "root.child-2",
							PartitionID: "1",
						},
					},
				},
				{
					ID:          "4",
					QueueName:   "root",
					PartitionID: "2",
					Children: []dao.PartitionQueueDAOInfo{
						{
							ID:          "5",
							QueueName:   "root.child-1",
							PartitionID: "2",
						},
						{
							ID:          "6",
							QueueName:   "root.child-2",
							PartitionID: "2",
						},
					},
				},
			},
			existingQueues: nil,
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "1",
						QueueName:   "root",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "2",
						QueueName:   "root.child-1",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "3",
						QueueName:   "root.child-2",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "4",
						QueueName:   "root",
						PartitionID: "2",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "5",
						QueueName:   "root.child-1",
						PartitionID: "2",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "6",
						QueueName:   "root.child-2",
						PartitionID: "2",
					},
				},
			},

			wantErr: false,
		},
		{
			name: "Sync queues with deeply nested queues",
			stateQueues: []dao.PartitionQueueDAOInfo{
				{
					ID:          "1",
					QueueName:   "root",
					PartitionID: "1",
					Children: []dao.PartitionQueueDAOInfo{
						{
							ID:          "2",
							QueueName:   "root.child-1",
							PartitionID: "1",
							Children: []dao.PartitionQueueDAOInfo{
								{
									ID:          "3",
									QueueName:   "root.child-1.1",
									PartitionID: "1",
									Children: []dao.PartitionQueueDAOInfo{
										{
											ID:          "4",
											QueueName:   "root.child-1.1.1",
											PartitionID: "1",
										},
										{
											ID:          "5",
											QueueName:   "root.child-1.1.2",
											PartitionID: "1",
										},
									},
								},
							},
						},
					},
				},
			},
			existingQueues: nil,
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "1",
						QueueName:   "root",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "2",
						QueueName:   "root.child-1",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "3",
						QueueName:   "root.child-1.1",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "4",
						QueueName:   "root.child-1.1.1",
						PartitionID: "1",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:          "5",
						QueueName:   "root.child-1.1.2",
						PartitionID: "1",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		ss.Run(tt.name, func() {
			// clean up the table after the test
			ss.T().Cleanup(func() {
				_, err := ss.pool.Exec(ctx, "DELETE FROM queues")
				require.NoError(ss.T(), err)
			})
			for _, q := range tt.existingQueues {
				err := ss.repo.InsertQueue(ctx, q)
				require.NoError(ss.T(), err)
			}

			s := NewService(ss.repo, nil, nil)

			err := s.syncQueues(context.Background(), tt.stateQueues)
			if tt.wantErr {
				require.Error(ss.T(), err)
				return
			}
			require.NoError(ss.T(), err)
			queuesInDB, err := s.repo.GetAllQueues(ctx)
			require.NoError(ss.T(), err)
			for _, target := range tt.expected {
				if !isQueuePresent(queuesInDB, target) {
					ss.T().Errorf("Queue %s in partition %s is not found in the DB", target.QueueName, target.PartitionID)
				}
			}
		})
	}
}

func isQueuePresent(queuesInDB []*model.Queue, targetQueue *model.Queue) bool {
	for _, dbQueue := range queuesInDB {
		if dbQueue.QueueName == targetQueue.QueueName && dbQueue.PartitionID == targetQueue.PartitionID {
			// Check if DeletedAtNano fields are either both nil or both non-nil
			if (dbQueue.DeletedAtNano == nil && targetQueue.DeletedAtNano != nil) ||
				(dbQueue.DeletedAtNano != nil && targetQueue.DeletedAtNano == nil) {
				return false // If one is nil and the other is not, return false
			}
			return true
		}
	}
	return false
}
