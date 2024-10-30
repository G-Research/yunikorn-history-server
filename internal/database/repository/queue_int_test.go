package repository

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/test/database"
)

func TestGetAllQueues_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	seedQueues(t, repo)

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
		t.Run(tt.name, func(t *testing.T) {
			queues, err := repo.GetAllQueues(context.Background())
			if err != nil {
				t.Fatalf("could not get queues: %v", err)
			}
			if len(queues) != tt.expectedTotalQueue {
				t.Fatalf("expected %d total queues, got %d", tt.expectedTotalQueue, len(queues))
			}
		})
	}
}

func TestGetQueuesInPartition_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	seedQueues(t, repo)

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
		t.Run(tt.name, func(t *testing.T) {
			queues, err := repo.GetQueuesInPartition(context.Background(), tt.partitionID)
			require.NoError(t, err)
			assert.Len(t, queues, tt.expectedTotalQueues)
		})
	}
}

func TestGetQueue_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	connPool := database.NewTestConnectionPool(ctx, t)
	repo, err := NewPostgresRepository(connPool)
	require.NoError(t, err, "could not create repository")

	seedQueues(t, repo)

	tests := []struct {
		name          string
		queueID       string
		partitionID   string
		expectedError bool
	}{
		{
			name:        "Get Queue root",
			queueID:     "1",
			partitionID: "1",
		},
		{
			name:        "Get Queue root.org.eng",
			queueID:     "3",
			partitionID: "1",
		},
		{
			name:          "Get non-existent Queue",
			queueID:       "99",
			partitionID:   "1",
			expectedError: true,
		},
		{
			name:        "Get Queue with no children",
			queueID:     "8",
			partitionID: "1",
		},
		{
			name:          "Get Queue from non-existent partition",
			queueID:       "1",
			partitionID:   "nonExistentPartition",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue, err := repo.GetQueueInPartition(context.Background(), tt.partitionID, tt.queueID)
			require.Equal(t, tt.expectedError, err != nil, "unexpected error", err)
			if tt.expectedError {
				return
			}
			assert.Equal(t, tt.queueID, queue.ID)
		})
	}
}

func TestDeleteQueues_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	seedQueues(t, repo)

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
		t.Run(tt.name, func(t *testing.T) {
			queues, err := repo.GetQueuesInPartition(context.Background(), tt.partitionID)
			if err != nil {
				t.Fatalf("could not get queues: %v", err)
			}
			now := time.Now()
			timestamp := now.UnixNano()
			for _, q := range queues {
				q.DeletedAtNano = &timestamp
				err := repo.UpdateQueue(ctx, q)
				require.NoError(t, err)
			}

			queues, err = repo.GetAllQueues(context.Background())
			if err != nil {
				t.Fatalf("could not get queues: %v", err)
			}
			// count the deleted queues
			var delQueues int
			for _, q := range queues {
				if q.DeletedAtNano != nil && q.PartitionID == tt.partitionID {
					delQueues++
				}
			}
			if delQueues != tt.expectedDelQueues {
				t.Fatalf("expected %d deleted queues, got %d", tt.expectedDelQueues, delQueues)
			}
		})
	}
}

func TestUpdateQueue_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	require.NoError(t, err)

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
						ID:              "1",
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
					ID:              "1",
					PartitionID:     "2",
					QueueName:       "root",
					CurrentPriority: 1,
				},
			},
			expectedError: false,
		},
		{
			name:           "Update root queue when root queue does not exist",
			existingQueues: nil,
			queueToUpdate: &model.Queue{
				Metadata: model.Metadata{
					CreatedAtNano: now.UnixNano(),
				},
				PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
					ID:              "2",
					PartitionID:     "1",
					QueueName:       "root",
					CurrentPriority: 1,
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clean up the table after the test
			t.Cleanup(func() {
				_, err := connPool.Exec(ctx, "DELETE FROM queues")
				require.NoError(t, err)
			})
			// seed the existing queues
			for _, q := range tt.existingQueues {
				if err := repo.InsertQueue(ctx, q); err != nil {
					t.Fatalf("could not seed queue: %v", err)
				}
			}
			// update the new queue
			err := repo.UpdateQueue(ctx, tt.queueToUpdate)
			if tt.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			// check if the queue is updated along with its children
			queueFromDB, err := repo.GetQueueInPartition(
				ctx,
				tt.queueToUpdate.PartitionID,
				tt.queueToUpdate.ID,
			)
			require.NoError(t, err)
			assert.Equal(t, tt.queueToUpdate.ID, queueFromDB.ID)
			assert.Equal(t, tt.queueToUpdate.QueueName, queueFromDB.QueueName)
			assert.Equal(t, tt.queueToUpdate.PartitionID, queueFromDB.PartitionID)
			assert.Equal(t, tt.queueToUpdate.CurrentPriority, queueFromDB.CurrentPriority)
			// compare the children
			for i, child := range tt.queueToUpdate.Children {
				assert.Equal(t, child.CurrentPriority, queueFromDB.Children[i].CurrentPriority)
			}
		})
	}
}

func seedQueues(t *testing.T, repo *PostgresRepository) {
	t.Helper()

	queues := []*model.Queue{
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "1",
				PartitionID: "1",
				QueueName:   "root",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "2",
				PartitionID: "1",
				QueueName:   "root.org",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "3",
				PartitionID: "1",
				QueueName:   "root.org.eng",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "4",
				PartitionID: "1",
				QueueName:   "root.org.eng.test",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "5",
				PartitionID: "1",
				QueueName:   "root.org.eng.prod",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "6",
				PartitionID: "1",
				QueueName:   "root.org.sales",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "7",
				PartitionID: "1",
				QueueName:   "root.org.sales.test",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "8",
				PartitionID: "1",
				QueueName:   "root.org.sales.prod",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "9",
				PartitionID: "1",
				QueueName:   "root.system",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "10",
				PartitionID: "2",
				QueueName:   "root",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "11",
				PartitionID: "2",
				QueueName:   "root.child",
			},
		},
		{
			Metadata: model.Metadata{
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				ID:          "12",
				PartitionID: "2",
				QueueName:   "root.child2",
			},
		},
	}

	for _, q := range queues {
		if err := repo.InsertQueue(context.Background(), q); err != nil {
			t.Fatalf("could not seed queue: %v", err)
		}
	}
}
