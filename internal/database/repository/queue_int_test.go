package repository

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type QueueTestSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *PostgresRepository
}

func (qs *QueueTestSuite) SetupSuite() {
	require.NotNil(qs.T(), qs.pool)
	repo, err := NewPostgresRepository(qs.pool)
	require.NoError(qs.T(), err)
	qs.repo = repo

	seedQueues(qs.T(), qs.repo)
}

func (qs *QueueTestSuite) TearDownSuite() {
	qs.pool.Close()
}

func (qs *QueueTestSuite) TestGetAllQueues() {
	ctx := context.Background()
	tests := []struct {
		name               string
		expectedTotalQueue int
	}{
		{
			name:               "Get All Queues",
			expectedTotalQueue: 12,
		},
	}

	for _, tt := range tests {
		qs.Run(tt.name, func() {
			queues, err := qs.repo.GetAllQueues(ctx)
			require.NoError(qs.T(), err)
			assert.Len(qs.T(), queues, tt.expectedTotalQueue)
		})
	}
}

func (qs *QueueTestSuite) TestGetQueuesInPartition() {
	ctx := context.Background()
	tests := []struct {
		name                string
		partitionID         string
		expectedTotalQueues int
	}{
		{
			name:                "Get Queues for default partition",
			partitionID:         "1",
			expectedTotalQueues: 9,
		},
		{
			name:                "Get Queues for second partition",
			partitionID:         "2",
			expectedTotalQueues: 3,
		},
	}

	for _, tt := range tests {
		qs.Run(tt.name, func() {
			queues, err := qs.repo.GetQueuesInPartition(ctx, tt.partitionID)
			require.NoError(qs.T(), err)
			assert.Len(qs.T(), queues, tt.expectedTotalQueues)
		})
	}
}

func (qs *QueueTestSuite) TestGetQueue() {
	ctx := context.Background()
	tests := []struct {
		name          string
		queueID       string
		expectedError bool
	}{
		{
			name:    "Get Queue root",
			queueID: "1",
		},
		{
			name:    "Get Queue root.org.eng",
			queueID: "3",
		},
		{
			name:          "Get non-existent Queue",
			queueID:       "99",
			expectedError: true,
		},
		{
			name:    "Get Queue with no children",
			queueID: "8",
		},
	}

	for _, tt := range tests {
		qs.Run(tt.name, func() {
			queue, err := qs.repo.GetQueue(ctx, tt.queueID)
			require.Equal(qs.T(), tt.expectedError, err != nil)
			if tt.expectedError {
				return
			}
			assert.Equal(qs.T(), tt.queueID, queue.ID)
		})
	}
}

func (qs *QueueTestSuite) TestDeleteQueues() {
	ctx := context.Background()
	tests := []struct {
		name              string
		partitionID       string
		expectedDelQueues int
	}{
		{
			name:              "Delete Queues for default partition",
			partitionID:       "1",
			expectedDelQueues: 9,
		},
		{
			name:              "Delete Queues for second partition",
			partitionID:       "2",
			expectedDelQueues: 3,
		},
	}

	for _, tt := range tests {
		qs.Run(tt.name, func() {
			queues, err := qs.repo.GetQueuesInPartition(ctx, tt.partitionID)
			require.NoError(qs.T(), err)
			now := time.Now()
			timestamp := now.UnixNano()
			for _, q := range queues {
				q.DeletedAtNano = &timestamp
				err := qs.repo.UpdateQueue(ctx, q)
				require.NoError(qs.T(), err)
			}

			queues, err = qs.repo.GetAllQueues(ctx)
			require.NoError(qs.T(), err)
			var delQueues int
			for _, q := range queues {
				if q.DeletedAtNano != nil && q.PartitionID == tt.partitionID {
					delQueues++
				}
			}
			assert.Equal(qs.T(), tt.expectedDelQueues, delQueues)
		})
	}
}

func (qs *QueueTestSuite) TestUpdateQueue() {
	ctx := context.Background()
	now := time.Now()
	tests := []struct {
		name           string
		existingQueues []*model.Queue
		queueToUpdate  *model.Queue
		expectedError  bool
	}{
		{
			name: "Update root queue when root queue exists",
			existingQueues: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now.UnixNano(),
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:              "13",
						PartitionID:     "1",
						QueueName:       "root",
						CurrentPriority: 0,
					},
				},
			},
			queueToUpdate: &model.Queue{
				Metadata: model.Metadata{
					CreatedAtNano: now.UnixNano(),
				},
				PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
					ID:              "13",
					PartitionID:     "2",
					QueueName:       "root",
					CurrentPriority: 1,
				},
			},
			expectedError: false,
		},
		{
			name:           "Update queue when queue does not exist",
			existingQueues: nil,
			queueToUpdate: &model.Queue{
				Metadata: model.Metadata{
					CreatedAtNano: now.UnixNano(),
				},
				PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
					ID:              "40",
					PartitionID:     "1",
					QueueName:       "non-existent-queue",
					CurrentPriority: 1,
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		qs.Run(tt.name, func() {
			for _, q := range tt.existingQueues {
				err := qs.repo.InsertQueue(ctx, q)
				require.NoError(qs.T(), err)
			}
			err := qs.repo.UpdateQueue(ctx, tt.queueToUpdate)
			require.Equal(qs.T(), tt.expectedError, err != nil)
			if tt.expectedError {
				return
			}
			queueFromDB, err := qs.repo.GetQueue(ctx, tt.queueToUpdate.ID)
			require.NoError(qs.T(), err)
			assert.Equal(qs.T(), tt.queueToUpdate.ID, queueFromDB.ID)
			assert.Equal(qs.T(), tt.queueToUpdate.QueueName, queueFromDB.QueueName)
			assert.Equal(qs.T(), tt.queueToUpdate.PartitionID, queueFromDB.PartitionID)
			assert.Equal(qs.T(), tt.queueToUpdate.CurrentPriority, queueFromDB.CurrentPriority)
			for i, child := range tt.queueToUpdate.Children {
				assert.Equal(qs.T(), child.CurrentPriority, queueFromDB.Children[i].CurrentPriority)
			}
		})
	}
}

func seedQueues(t *testing.T, repo *PostgresRepository) {
	t.Helper()

	now := time.Now().UnixNano()
	queueData := []struct {
		ID          string
		PartitionID string
		QueueName   string
	}{
		{"1", "1", "root"},
		{"2", "1", "root.org"},
		{"3", "1", "root.org.eng"},
		{"4", "1", "root.org.eng.test"},
		{"5", "1", "root.org.eng.prod"},
		{"6", "1", "root.org.sales"},
		{"7", "1", "root.org.sales.test"},
		{"8", "1", "root.org.sales.prod"},
		{"9", "1", "root.system"},
		{"10", "2", "root"},
		{"11", "2", "root.child"},
		{"12", "2", "root.child2"},
	}

	for _, qd := range queueData {
		queue := &model.Queue{
			Metadata: model.Metadata{
				CreatedAtNano: now,
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          qd.ID,
				PartitionID: qd.PartitionID,
				QueueName:   qd.QueueName,
			},
		}
		if err := repo.InsertQueue(context.Background(), queue); err != nil {
			t.Fatalf("could not seed queue %s: %v", qd.QueueName, err)
		}
	}
}
