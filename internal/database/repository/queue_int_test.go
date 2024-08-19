package repository

import (
	"context"
	"testing"

	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/test/database"
	"github.com/apache/yunikorn-core/pkg/webservice/dao"
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
			expectedTotalQueue: 11,
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
			expectedTotalQueue: 2,
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
			QueueName: "root.child",
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
			queues = append(queues, flattenQueues(util.ToPtrSlice(q.Children))...)
		}
	}
	return queues
}
