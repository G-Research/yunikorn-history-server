package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/util"

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

func TestGetQueuesPerPartition_Integration(t *testing.T) {
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
		partition          string
		expectedRootQueue  int
		expectedTotalQueue int
	}{
		{
			name:               "Get Queues for default partition",
			partition:          "default",
			expectedRootQueue:  1,
			expectedTotalQueue: 9,
		},
		{
			name:               "Get Queues for second partition",
			partition:          "second",
			expectedRootQueue:  1,
			expectedTotalQueue: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queues, err := repo.GetQueuesPerPartition(context.Background(), tt.partition)
			if err != nil {
				t.Fatalf("could not get queues: %v", err)
			}
			if len(queues) != tt.expectedRootQueue {
				t.Fatalf("expected %d root queues, got %d", tt.expectedRootQueue, len(queues))
			}
			queues = flattenQueues(queues)
			if len(queues) != tt.expectedTotalQueue {
				t.Fatalf("expected %d total queues, got %d", tt.expectedTotalQueue, len(queues))
			}
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
			queue, err := repo.GetQueue(context.Background(), tt.partition, tt.queueName)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.queueName, queue.QueueName)

				queues := flattenQueues(util.ToPtrSlice(queue.Children))
				assert.Equal(t, tt.expectedChildrenQueueLen, len(queues))
			}
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
			queues, err := repo.GetQueuesPerPartition(context.Background(), tt.partition)
			if err != nil {
				t.Fatalf("could not get queues: %v", err)
			}
			if err := repo.DeleteQueues(ctx, queues); err != nil {
				t.Fatalf("could not delete queues: %v", err)
			}
			queues, err = repo.GetAllQueues(context.Background())
			if err != nil {
				t.Fatalf("could not get queues: %v", err)
			}
			// count the deleted queues
			var delQueues int
			for _, q := range queues {
				if q.DeletedAt.Valid && q.Partition == tt.partition {
					delQueues++
				}
			}
			if delQueues != tt.expectedDelQueues {
				t.Fatalf("expected %d deleted queues, got %d", tt.expectedDelQueues, delQueues)
			}
		})
	}
}

func seedQueues(t *testing.T, repo *PostgresRepository) {
	t.Helper()

	queues := []*dao.PartitionQueueDAOInfo{
		{
			Partition: "default",
			QueueName: "root",
		},
		{
			Partition: "default",
			QueueName: "root.org",
			Parent:    "root",
		},
		{
			Partition: "default",
			QueueName: "root.system",
			Parent:    "root",
			IsLeaf:    true,
		},
		{
			Partition: "default",
			QueueName: "root.org.eng",
			Parent:    "root.org",
		},
		{
			Partition: "default",
			QueueName: "root.org.eng.test",
			Parent:    "root.org.eng",
			IsLeaf:    true,
		},
		{
			Partition: "default",
			QueueName: "root.org.eng.prod",
			Parent:    "root.org.eng",
			IsLeaf:    true,
		},
		{
			Partition: "default",
			QueueName: "root.org.sales",
			Parent:    "root.org",
		},
		{
			Partition: "default",
			QueueName: "root.org.sales.test",
			Parent:    "root.org.sales",
			IsLeaf:    true,
		},
		{
			Partition: "default",
			QueueName: "root.org.sales.prod",
			Parent:    "root.org.sales",
			IsLeaf:    true,
		},
		{
			Partition: "second",
			QueueName: "root",
		},
		{
			Partition: "second",
			QueueName: "root.child1",
			Parent:    "root",
			IsLeaf:    true,
		},
		{
			Partition: "second",
			QueueName: "root.child2",
			Parent:    "root",
			IsLeaf:    true,
		},
	}

	if err := repo.UpsertQueues(context.Background(), queues); err != nil {
		t.Fatalf("could not seed queue: %v", err)
	}
}

func flattenQueues(qs []*model.PartitionQueueDAOInfo) []*model.PartitionQueueDAOInfo {
	var queues []*model.PartitionQueueDAOInfo
	for _, q := range qs {
		queues = append(queues, q)
		if len(q.Children) > 0 {
			queues = append(queues, flattenQueues(q.Children)...)
		}
	}
	return queues
}
