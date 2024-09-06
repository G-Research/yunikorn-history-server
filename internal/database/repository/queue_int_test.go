package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
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

				queues := flattenQueues(queue.Children)
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

func TestAddQueues_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	connPool := database.NewTestConnectionPool(ctx, t)

	repo, err := NewPostgresRepository(connPool)
	if err != nil {
		t.Fatalf("could not create repository: %v", err)
	}

	tests := []struct {
		name           string
		existingQueues []*dao.PartitionQueueDAOInfo
		newQueues      []*dao.PartitionQueueDAOInfo
		expectedError  error
	}{
		{
			name:           "Add Queues in empty DB",
			existingQueues: nil,
			newQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "default",
							QueueName: "root.org",
							Parent:    "root",
							IsLeaf:    true,
						},
						{
							Partition: "default",
							QueueName: "root.system",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
		},
		{
			name: "Add Queues when non conflicting queues exist in same partition",
			existingQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "secondRoot",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "default",
							QueueName: "secondRoot.child1",
							Parent:    "root",
							IsLeaf:    true,
						},
						{
							Partition: "default",
							QueueName: "secondRoot.child2",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
			newQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "default",
							QueueName: "root.org",
							Parent:    "root",
							IsLeaf:    true,
						},
						{
							Partition: "default",
							QueueName: "root.system",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
		},
		{
			name: "Add Queues when conflicting queues exist in different partition",
			existingQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "secondPartition",
					QueueName: "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "secondPartition",
							QueueName: "root.org",
							Parent:    "root",
							IsLeaf:    true,
						},
						{
							Partition: "secondPartition",
							QueueName: "root.system",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
			newQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "default",
							QueueName: "root.org",
							Parent:    "root",
							IsLeaf:    true,
						},
						{
							Partition: "default",
							QueueName: "root.system",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
		},
		{
			name: "Add Queues when conflicting root queues exist in same partition",
			existingQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "default",
							QueueName: "root.org",
							Parent:    "root",
							IsLeaf:    true,
						},
						{
							Partition: "default",
							QueueName: "root.system",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
			newQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "default",
							QueueName: "root.org",
							Parent:    "root",
							IsLeaf:    true,
						},
						{
							Partition: "default",
							QueueName: "root.system",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
			expectedError: fmt.Errorf("could not add queue %s into DB", "root"),
		},
		// Add Queues when conflicting child queues exist in same partition under different parent
		// Sudipto: This test case is not valid as the parent queueName is appended to the child queueName
		// child queueName = parentName.childName
		// {...},
		{
			// AddQueues() will use partition name from the parent queue to set the partition name for the child queue
			name:           "Add Queues when partition name is empty for child queue",
			existingQueues: nil,
			newQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "", // yunikorn returns empty string for child queue partition
							QueueName: "root.org",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
		},
		{
			// We should not encounter this case as yunikorn should not return empty string for partition name for a root queue
			name:           "Add Queues when partition name is empty for root queue",
			existingQueues: nil,
			newQueues: []*dao.PartitionQueueDAOInfo{
				{
					Partition: "",
					QueueName: "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "", // yunikorn returns empty string for child queue partition
							QueueName: "root.org",
							Parent:    "root",
							IsLeaf:    true,
						},
					},
				},
			},
			expectedError: fmt.Errorf("partition is required for queue %s", "root"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clean up the table after the test
			defer func() {
				_, err := connPool.Exec(ctx, "DELETE FROM queues")
				if err != nil {
					t.Fatalf("could not empty queue table: %v", err)
				}
			}()
			// seed the existing queues
			if tt.existingQueues != nil {
				if err := repo.AddQueues(ctx, nil, tt.existingQueues); err != nil {
					t.Fatalf("could not seed queue: %v", err)
				}
			}
			// add the new queues
			err := repo.AddQueues(ctx, nil, tt.newQueues)
			if tt.expectedError != nil {
				require.Error(t, err)
				// match expected error message exist in the actual error message
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				for _, q := range tt.newQueues {
					// check if the queue is added along with its children
					queueFromDB, err := repo.GetQueue(ctx, q.Partition, q.QueueName)
					require.NoError(t, err)
					assert.Equal(t, q.QueueName, queueFromDB.QueueName)
					assert.Equal(t, q.Partition, queueFromDB.Partition)
					assert.Equal(t, len(q.Children), len(queueFromDB.Children))
				}

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
			Children: []dao.PartitionQueueDAOInfo{
				{
					Partition: "default",
					QueueName: "root.org",
					Parent:    "root",
					Children: []dao.PartitionQueueDAOInfo{
						{
							Partition: "default",
							QueueName: "root.org.eng",
							Parent:    "root.org",
							Children: []dao.PartitionQueueDAOInfo{
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
							},
						},
						{
							Partition: "default",
							QueueName: "root.org.sales",
							Parent:    "root.org",
							Children: []dao.PartitionQueueDAOInfo{
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
							},
						},
					},
				},
				{
					Partition: "default",
					QueueName: "root.system",
					Parent:    "root",
					IsLeaf:    true,
				},
			},
		},
		{
			Partition: "second",
			QueueName: "root",
			Children: []dao.PartitionQueueDAOInfo{
				{
					Partition: "second",
					QueueName: "root.child",
					Parent:    "root",
					IsLeaf:    true,
				},
				{
					Partition: "second",
					QueueName: "root.child2",
					Parent:    "root",
					IsLeaf:    true,
				},
			},
		},
	}

	if err := repo.AddQueues(context.Background(), nil, queues); err != nil {
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
