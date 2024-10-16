package repository

import (
	"context"
	"fmt"
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
		name                     string
		queueName                string
		partition                string
		expectedChildrenQueueLen int // all children in the tree
		expectedError            error
	}{
		{
			name:                     "Get Queue root",
			queueName:                "root",
			partition:                "default",
			expectedChildrenQueueLen: 8,
			expectedError:            nil,
		},
		{
			name:                     "Get Queue root.org.eng",
			queueName:                "root.org.eng",
			partition:                "default",
			expectedChildrenQueueLen: 2,
			expectedError:            nil,
		},
		{
			name:          "Get non-existent Queue",
			queueName:     "non-existent",
			partition:     "default",
			expectedError: fmt.Errorf("queue not found: %s", "non-existent"),
		},
		{
			name:                     "Get Queue with no children",
			queueName:                "root.org.sales.prod",
			partition:                "default",
			expectedChildrenQueueLen: 0,
			expectedError:            nil,
		},
		{
			name:          "Get deleted Queue",
			queueName:     "deletedQueue",
			partition:     "default",
			expectedError: fmt.Errorf("queue not found: %s", "deletedQueue"),
		},
		{
			name:          "Get Queue from non-existent partition",
			queueName:     "root",
			partition:     "nonExistentPartition",
			expectedError: fmt.Errorf("queue not found: %s", "root"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue, err := repo.GetQueueInPartition(context.Background(), tt.partition, tt.queueName)
			require.Equal(t, tt.expectedError, err != nil, "unexpected error", err)

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
				q.DeletedAt = &timestamp
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
				if q.DeletedAt != nil && q.Partition == tt.partition {
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

					ModelMetadata: model.ModelMetadata{
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
				ModelMetadata: model.ModelMetadata{
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
				ModelMetadata: model.ModelMetadata{
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
			ModelMetadata: model.ModelMetadata{
				ID:        "1",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "2",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "3",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.eng",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "4",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.eng.test",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "5",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.eng.prod",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "6",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.sales",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "7",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.sales.test",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "8",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.org.sales.prod",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "9",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "default",
				QueueName: "root.system",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "10",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "second",
				QueueName: "root",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "11",
				CreatedAt: time.Now().UnixNano(),
			},
			PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
				Partition: "second",
				QueueName: "root.child",
			},
		},
		{
			ModelMetadata: model.ModelMetadata{
				ID:        "12",
				CreatedAt: time.Now().UnixNano(),
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
