package yunikorn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/database/migrations"
	"github.com/G-Research/yunikorn-history-server/internal/database/postgres"
	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/test/config"
	"github.com/G-Research/yunikorn-history-server/test/database"
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
		setup         func() *httptest.Server
		partitions    []*model.Partition
		existingNodes []*model.Node
		expectedNodes []*model.Node
		wantErr       bool
	}{
		{
			name: "Sync nodes with no existing nodes",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					partitionName := extractPartitionNameFromURL(r.URL.Path)
					response := []dao.NodeDAOInfo{
						{ID: "1", NodeID: "node-1", Partition: partitionName, HostName: "host-1"},
						{ID: "2", NodeID: "node-2", Partition: partitionName, HostName: "host-2"},
					}
					writeResponse(t, w, response)
				}))
			},
			partitions: []*model.Partition{
				{
					PartitionInfo: dao.PartitionInfo{
						Name: "default",
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
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					partitionName := extractPartitionNameFromURL(r.URL.Path)
					response := []dao.NodeDAOInfo{
						{ID: "2", NodeID: "node-2", Partition: partitionName, HostName: "host-2-updated"},
					}
					writeResponse(t, w, response)
				}))
			},
			partitions: []*model.Partition{
				{
					PartitionInfo: dao.PartitionInfo{
						Name: "default",
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

			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))
			s := NewService(repo, eventRepository, client)

			err := s.syncNodes(ctx, tt.partitions)
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
		setup          func() *httptest.Server
		partitions     []*model.Partition
		existingQueues []*model.Queue
		expected       []*model.Queue
		wantErr        bool
	}{
		{
			name: "Sync queues with no existing queues",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					partitionName := extractPartitionNameFromURL(r.URL.Path)
					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: partitionName,
						Children: []dao.PartitionQueueDAOInfo{
							{
								QueueName: "root.child-1",
								Children: []dao.PartitionQueueDAOInfo{
									{QueueName: "root.child-1.1"},
									{QueueName: "root.child-1.2"},
								},
							},
						},
					}
					writeResponse(t, w, response)
				}))
			},
			partitions: []*model.Partition{
				{
					PartitionInfo: dao.PartitionInfo{
						Name: "default",
					},
				},
			},
			existingQueues: nil,
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1.1",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1.2",
						Partition: "default",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Sync queues with existing queues in DB",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					partitionName := extractPartitionNameFromURL(r.URL.Path)

					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: partitionName,
						Children: []dao.PartitionQueueDAOInfo{
							{
								QueueName: "root.child-1",
							},
							{
								QueueName: "root.child-2",
							},
						},
					}
					writeResponse(t, w, response)
				}))
			},
			partitions: []*model.Partition{
				{
					PartitionInfo: dao.PartitionInfo{
						Name: "default",
					},
				},
			},
			existingQueues: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "1",
						QueueName: "root",
						Partition: "default",
					},
				},
			},
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-2",
						Partition: "default",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Sync queues when queue is deleted",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					partitionName := extractPartitionNameFromURL(r.URL.Path)
					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: partitionName,
						Children: []dao.PartitionQueueDAOInfo{
							{
								QueueName: "root.child-2",
							},
						},
					}
					writeResponse(t, w, response)
				}))
			},
			partitions: []*model.Partition{
				{
					PartitionInfo: dao.PartitionInfo{
						Name: "default",
					},
				},
			},
			existingQueues: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "1",
						QueueName: "root",
						Partition: "default",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "2",
						QueueName: "root.child-1",
						Partition: "default",
					},
				},
			},
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-2",
						Partition: "default",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Sync queues with HTTP error",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}))
			},
			partitions: []*model.Partition{{
				PartitionInfo: dao.PartitionInfo{
					Name: "default",
				},
			}},
			existingQueues: nil,
			expected:       nil,
			wantErr:        true,
		},
		{
			name: "Sync queues with multiple partitions",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					partitionName := extractPartitionNameFromURL(r.URL.Path)
					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: partitionName,
						Children: []dao.PartitionQueueDAOInfo{
							{QueueName: "root.child-1"},
							{QueueName: "root.child-2"},
						},
					}
					writeResponse(t, w, response)
				}))
			},
			partitions: []*model.Partition{
				{
					PartitionInfo: dao.PartitionInfo{
						Name: "default",
					},
				},
				{
					PartitionInfo: dao.PartitionInfo{
						Name: "secondary",
					},
				},
				{
					PartitionInfo: dao.PartitionInfo{
						Name: "third",
					},
				},
			},
			existingQueues: nil,
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-2",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: "secondary",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1",
						Partition: "secondary",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-2",
						Partition: "secondary",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: "third",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1",
						Partition: "third",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-2",
						Partition: "third",
					},
				},
			},

			wantErr: false,
		},
		{
			name: "Sync queues with deeply nested queues",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					partitionName := extractPartitionNameFromURL(r.URL.Path)

					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: partitionName,
						Children: []dao.PartitionQueueDAOInfo{
							{
								QueueName: "root.child-1",
								Children: []dao.PartitionQueueDAOInfo{
									{
										QueueName: "root.child-1.1",
										Children: []dao.PartitionQueueDAOInfo{
											{QueueName: "root.child-1.1.1"},
											{QueueName: "root.child-1.1.2"},
										},
									},
								},
							},
						},
					}
					writeResponse(t, w, response)
				}))
			},
			partitions: []*model.Partition{{
				PartitionInfo: dao.PartitionInfo{
					Name: "default",
				},
			}},
			existingQueues: nil,
			expected: []*model.Queue{
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1.1",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1.1.1",
						Partition: "default",
					},
				},
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1.1.2",
						Partition: "default",
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

			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))
			s := NewService(repo, eventRepository, client)

			// Create a cancellable context for this specific service
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Start the service in a goroutine
			go func() {
				_ = s.Run(ctx)
			}()

			err := s.syncQueues(context.Background(), tt.partitions)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			queuesInDB, err := s.repo.GetAllQueues(ctx)
			require.NoError(t, err)
			for _, target := range tt.expected {
				if !isQueuePresent(queuesInDB, target) {
					t.Errorf("Queue %s in partition %s is not found in the DB", target.QueueName, target.Partition)
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
		setup              func() *httptest.Server
		existingPartitions []*model.Partition
		expected           []*model.Partition
		wantErr            bool
	}{
		{
			name: "Sync partition with no existing partitions in DB",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.PartitionInfo{
						{Name: "default"},
						{Name: "secondary"},
					}
					writeResponse(t, w, response)
				}))
			},
			existingPartitions: nil,
			expected: []*model.Partition{
				{
					PartitionInfo: dao.PartitionInfo{
						ID:   "1",
						Name: "default",
					},
				},
				{
					PartitionInfo: dao.PartitionInfo{
						ID:   "2",
						Name: "secondary",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should mark secondary partition as deleted in DB",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.PartitionInfo{
						{
							ID:   "1",
							Name: "default",
						},
					}
					writeResponse(t, w, response)
				}))
			},
			existingPartitions: []*model.Partition{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:   "1",
						Name: "default",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:   "2",
						Name: "secondary",
					},
				},
			},
			expected: []*model.Partition{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionInfo: dao.PartitionInfo{
						ID:   "3",
						Name: "default",
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
				_, err := pool.Exec(ctx, "DELETE FROM partitions")
				require.NoError(t, err)
			})

			for _, partition := range tt.existingPartitions {
				err := repo.InsertPartition(ctx, partition)
				require.NoError(t, err)
			}

			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))
			s := NewService(repo, eventRepository, client)

			// Start the service
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = s.Run(ctx)
			}()

			partitions, err := s.syncPartitions(ctx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			sort.Slice(partitions, func(i, j int) bool {
				return partitions[i].Name < partitions[j].Name
			})

			var partitionsInDB []*model.Partition
			partitionsInDB, err = s.repo.GetAllPartitions(ctx, repository.PartitionFilters{})
			require.NoError(t, err)

			sort.Slice(partitionsInDB, func(i, j int) bool {
				return partitionsInDB[i].Name < partitionsInDB[j].Name
			})

			i := 0
			j := 0
			for i < len(partitions) && j < len(partitionsInDB) {
				newPartition := partitions[i]
				dbPartition := partitionsInDB[j]
				if newPartition.ID == dbPartition.ID {
					assert.Equal(t, newPartition.PartitionInfo, dbPartition.PartitionInfo)
					assert.Nil(t, newPartition.DeletedAtNano)
					i++
					j++
					continue
				}
				assert.NotNil(t, dbPartition.DeletedAtNano)
				j++
			}
			assert.Equal(t, i, len(partitions))

			assert.Equal(t, len(partitions), i)
			for i := j; i < len(partitionsInDB); i++ {
				assert.NotNil(t, partitionsInDB[i].DeletedAtNano)
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
		setup                func() *httptest.Server
		existingApplications []*model.Application
		expectedLive         []*model.Application
		expectedDeleted      []*model.Application
		wantErr              bool
	}{
		{
			name: "Sync applications with no existing applications in DB",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.ApplicationDAOInfo{
						{ID: "1", ApplicationID: "app-1"},
						{ID: "2", ApplicationID: "app-2"},
					}
					writeResponse(t, w, response)
				}))
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
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.ApplicationDAOInfo{
						{ID: "1", ApplicationID: "app-1"},
					}
					writeResponse(t, w, response)
				}))
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

			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))
			s := NewService(repo, eventRepository, client)

			// Start the service
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = s.Run(ctx)
			}()

			err := s.syncApplications(ctx)
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
		if dbQueue.QueueName == targetQueue.QueueName && dbQueue.Partition == targetQueue.Partition {
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

// Helper function to extract partition name from the URL
func extractPartitionNameFromURL(urlPath string) string {
	// Assume URL is like: /ws/v1/partition/{partitionName}/queues
	parts := strings.Split(urlPath, "/")
	if len(parts) > 4 {
		return parts[4]
	}
	return ""
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
		database.DropTestSchema(ctx, t, schema)
	}

	return pool, repo, cleanup
}
