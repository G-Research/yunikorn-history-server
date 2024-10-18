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
		partition           string
		expectedTotalQueues int
	}{
		{
			name:                "Get Queues for default partition",
			partition:           "default",
			expectedTotalQueues: 9,
		},
		{
			name:                "Get Queues for second partition",
			partition:           "second",
			expectedTotalQueues: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queues, err := repo.GetQueuesInPartition(context.Background(), tt.partition)
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
		queueName     string
		partition     string
		expectedError bool
	}{
		{
			name:      "Get Queue root",
			queueName: "root",
			partition: "default",
		},
		{
			name:      "Get Queue root.org.eng",
			queueName: "root.org.eng",
			partition: "default",
		},
		{
			name:          "Get non-existent Queue",
			queueName:     "non-existent",
			partition:     "default",
			expectedError: true,
		},
		{
			name:      "Get Queue with no children",
			queueName: "root.org.sales.prod",
			partition: "default",
		},
		{
			name:          "Get deleted Queue",
			queueName:     "deletedQueue",
			partition:     "default",
			expectedError: true,
		},
		{
			name:          "Get Queue from non-existent partition",
			queueName:     "root",
			partition:     "nonExistentPartition",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue, err := repo.GetQueueInPartition(context.Background(), tt.partition, tt.queueName)
			require.Equal(t, tt.expectedError, err != nil, "unexpected error", err)
			if tt.expectedError {
				return
			}
			assert.Equal(t, tt.queueName, queue.QueueName)
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
		partition         string
		expectedDelQueues int
	}{
		{
			name:              "Delete Queues for default partition",
			partition:         "default",
			expectedDelQueues: 9,
		},
		{
			name:              "Delete Queues for second partition",
			partition:         "second",
			expectedDelQueues: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queues, err := repo.GetQueuesInPartition(context.Background(), tt.partition)
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
				if q.DeletedAtNano != nil && q.Partition == tt.partition {
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
						ID: "1",
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						Partition:       "default",
						QueueName:       "root",
						CurrentPriority: 0,
					},
				},
			},
			queueToUpdate: &model.Queue{
				Metadata: model.Metadata{
					ID: "1",
				},
				PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
					Partition:       "default",
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
					ID: "2",
				},
				PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
					Partition:       "default",
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
			queueFromDB, err := repo.GetQueueInPartition(ctx, tt.queueToUpdate.Partition, tt.queueToUpdate.QueueName)
			require.NoError(t, err)
			assert.Equal(t, tt.queueToUpdate.QueueName, queueFromDB.QueueName)
			assert.Equal(t, tt.queueToUpdate.Partition, queueFromDB.Partition)
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
				ID:            "1",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "2",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "3",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.eng",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "4",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.eng.test",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "5",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.eng.prod",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "6",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.sales",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "7",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.sales.test",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "8",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.sales.prod",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "9",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.system",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "10",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "second",
				QueueName: "root",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "11",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "second",
				QueueName: "root.child",
			},
		},
		{
			Metadata: model.Metadata{
				ID:            "12",
				CreatedAtNano: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "second",
				QueueName: "root.child2",
			},
		},
	}

	for _, q := range queues {
		if err := repo.InsertQueue(context.Background(), q); err != nil {
			t.Fatalf("could not seed queue: %v", err)
		}
	}
}
