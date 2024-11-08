package yunikorn

import (
	"context"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/unicorn-history-server/internal/database/migrations"
	"github.com/G-Research/unicorn-history-server/internal/database/postgres"
	"github.com/G-Research/unicorn-history-server/internal/database/repository"
	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/unicorn-history-server/test/config"
	"github.com/G-Research/unicorn-history-server/test/database"
)

func TestSync_syncNodes_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	pool, repo, cleanupDB := setupDatabase(t, ctx)
	t.Cleanup(cleanupDB)
	eventRepository := repository.NewInMemoryEventRepository()

	nowNano := time.Now().UnixNano()

	tests := []struct {
		name          string
		stateNodes    []*dao.NodesDAOInfo
		existingNodes []*model.Node
		expectedNodes []*model.Node
		wantErr       bool
	}{
		{
			name: "Sync nodes with no existing nodes",
			stateNodes: []*dao.NodesDAOInfo{
				{
					PartitionName: "default",
					Nodes: []*dao.NodeDAOInfo{
						{ID: "1", NodeID: "node-1", Partition: "default", HostName: "host-1"},
						{ID: "2", NodeID: "node-2", Partition: "default", HostName: "host-2"},
					},
				},
			},
			existingNodes: nil,
			expectedNodes: []*model.Node{
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "2", NodeID: "node-2", HostName: "host-2"}},
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "1", NodeID: "node-1", HostName: "host-1"}},
			},
			wantErr: false,
		},
		{
			name: "Sync nodes with existing nodes in DB",
			stateNodes: []*dao.NodesDAOInfo{
				{
					PartitionName: "default",
					Nodes: []*dao.NodeDAOInfo{
						{ID: "2", NodeID: "node-2", Partition: "default", HostName: "host-2-updated"},
					},
				},
			},
			existingNodes: []*model.Node{
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "1", NodeID: "node-1", HostName: "host-1", Partition: "default"}},
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "2", NodeID: "node-2", HostName: "host-2", Partition: "default"}},
			},
			expectedNodes: []*model.Node{
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "2", NodeID: "node-2", HostName: "host-2-updated", Partition: "default"}}, // updated
				{
					Metadata:    model.Metadata{DeletedAtNano: &nowNano}, // deleted
					NodeDAOInfo: dao.NodeDAOInfo{ID: "1", NodeID: "node-1", HostName: "host-1", Partition: "default"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clean up the table after the test
			t.Cleanup(func() {
				_, err := pool.Exec(ctx, "DELETE FROM nodes")
				require.NoError(t, err)
			})

			for _, node := range tt.existingNodes {
				err := repo.InsertNode(ctx, node)
				require.NoError(t, err)
			}

			s := NewService(repo, eventRepository, nil)

			err := s.syncNodes(ctx, tt.stateNodes)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			nodesInDB, err := s.repo.GetNodesPerPartition(ctx, "default", repository.NodeFilters{})
			require.NoError(t, err)
			for i, target := range tt.expectedNodes {
				require.Equal(t, target.ID, nodesInDB[i].ID)
				require.Equal(t, target.NodeID, nodesInDB[i].NodeID)
				require.Equal(t, target.HostName, nodesInDB[i].HostName)
				if target.DeletedAtNano != nil {
					require.NotNil(t, nodesInDB[i].DeletedAtNano)
				}
			}
		})
	}
}

func TestSync_syncQueues_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	pool, repo, cleanupDB := setupDatabase(t, ctx)
	t.Cleanup(cleanupDB)
	eventRepository := repository.NewInMemoryEventRepository()

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
		t.Run(tt.name, func(t *testing.T) {
			// clean up the table after the test
			t.Cleanup(func() {
				_, err := pool.Exec(ctx, "DELETE FROM queues")
				require.NoError(t, err)
			})

			for _, q := range tt.existingQueues {
				err := repo.InsertQueue(ctx, q)
				require.NoError(t, err)
			}

			s := NewService(repo, eventRepository, nil)

			err := s.syncQueues(context.Background(), tt.stateQueues)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			queuesInDB, err := s.repo.GetAllQueues(ctx)
			require.NoError(t, err)
			for _, target := range tt.expected {
				if !isQueuePresent(queuesInDB, target) {
					t.Errorf("Queue %s in partition %s is not found in the DB", target.QueueName, target.PartitionID)
				}
			}
		})
	}
}

func TestSync_syncPartitions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, repo, cleanupDB := setupDatabase(t, ctx)
	t.Cleanup(cleanupDB)
	eventRepository := repository.NewInMemoryEventRepository()

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
		t.Run(tt.name, func(t *testing.T) {
			// clean up the table after the test
			t.Cleanup(func() {
				_, err := pool.Exec(ctx, "DELETE FROM partitions")
				require.NoError(t, err)
			})

			for _, partition := range tt.existingPartitions {
				err := repo.InsertPartition(ctx, partition)
				require.NoError(t, err)
			}

			s := NewService(repo, eventRepository, nil)

			// Start the service
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := s.syncPartitions(ctx, tt.statePartitions)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var partitionsInDB []*model.Partition
			partitionsInDB, err = s.repo.GetAllPartitions(ctx, repository.PartitionFilters{})
			require.NoError(t, err)

			for _, dbPartition := range partitionsInDB {
				found := false
				for _, expectedPartition := range tt.expected {
					if dbPartition.ID == expectedPartition.ID {
						assert.Equal(t, expectedPartition.PartitionInfo, dbPartition.PartitionInfo)
						assert.Nil(t, expectedPartition.DeletedAtNano)
						found = true
					}
				}
				if !found {
					assert.NotNil(t, dbPartition.DeletedAtNano)
				}
			}

		})
	}
}

func TestSync_syncApplications_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	pool, repo, cleanupDB := setupDatabase(t, ctx)
	t.Cleanup(cleanupDB)
	eventRepository := repository.NewInMemoryEventRepository()

	now := time.Now().UnixNano()
	tests := []struct {
		name                 string
		stateApplications    []*dao.ApplicationDAOInfo
		existingApplications []*model.Application
		expectedLive         []*model.Application
		expectedDeleted      []*model.Application
		wantErr              bool
	}{
		{
			name: "Sync applications with no existing applications in DB",
			stateApplications: []*dao.ApplicationDAOInfo{
				{ID: "1", ApplicationID: "app-1"},
				{ID: "2", ApplicationID: "app-2"},
			},
			existingApplications: nil,
			expectedLive: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "1",
						ApplicationID: "app-1",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "2",
						ApplicationID: "app-2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should mark application as deleted in DB",
			stateApplications: []*dao.ApplicationDAOInfo{
				{ID: "1", ApplicationID: "app-1"},
			},
			existingApplications: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "1",
						ApplicationID: "app-1",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "2",
						ApplicationID: "app-2",
					},
				},
			},
			expectedLive: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "1",
						ApplicationID: "app-1",
					},
				},
			},
			expectedDeleted: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "2",
						ApplicationID: "app-2",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clean up the table after the test
			t.Cleanup(func() {
				_, err := pool.Exec(ctx, "DELETE FROM applications")
				require.NoError(t, err)
			})

			for _, app := range tt.existingApplications {
				err := repo.InsertApplication(ctx, app)
				require.NoError(t, err)
			}

			s := NewService(repo, eventRepository, nil)

			// Start the service
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := s.syncApplications(ctx, tt.stateApplications)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			applicationsInDB, err := s.repo.GetAllApplications(
				ctx,
				repository.ApplicationFilters{},
			)
			require.NoError(t, err)

			require.Equal(t, len(tt.expectedLive)+len(tt.expectedDeleted), len(applicationsInDB))

			lookup := make(map[string]model.Application)
			for _, app := range applicationsInDB {
				lookup[app.ID] = *app
			}

			for _, target := range tt.expectedLive {
				state, ok := lookup[target.ID]
				require.True(t, ok)
				assert.NotEmpty(t, state.ID)
				assert.Greater(t, state.Metadata.CreatedAtNano, int64(0))
				assert.Nil(t, state.Metadata.DeletedAtNano)
			}

			for _, target := range tt.expectedDeleted {
				state, ok := lookup[target.ID]
				require.True(t, ok)
				assert.NotEmpty(t, state.ID)
				assert.Greater(t, state.Metadata.CreatedAtNano, int64(0))
				assert.NotNil(t, state.Metadata.DeletedAtNano)
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

func setupDatabase(t *testing.T, ctx context.Context) (*pgxpool.Pool, repository.Repository, func()) {
	schema := database.CreateTestSchema(ctx, t)
	cfg := config.GetTestPostgresConfig()
	cfg.Schema = schema
	m, err := migrations.New(cfg, "../../migrations")
	if err != nil {
		t.Fatalf("error creating migrator: %v", err)
	}
	applied, err := m.Up()
	if err != nil {
		t.Fatalf("error occured while applying migrations: %v", err)
	}
	if !applied {
		t.Fatal("migrator finished but migrations were not applied")
	}

	pool, err := postgres.NewConnectionPool(ctx, cfg)
	if err != nil {
		t.Fatalf("error creating postgres connection pool: %v", err)
	}
	repo, err := repository.NewPostgresRepository(pool)
	if err != nil {
		t.Fatalf("error creating postgres repository: %v", err)
	}

	cleanup := func() {
		database.DropTestSchema(context.Background(), t, schema)
	}

	return pool, repo, cleanup
}
