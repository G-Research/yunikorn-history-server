package yunikorn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/model"

	"github.com/G-Research/yunikorn-history-server/internal/database/migrations"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/yunikorn-history-server/internal/database/postgres"
	"github.com/G-Research/yunikorn-history-server/internal/database/repository"

	"github.com/G-Research/yunikorn-history-server/test/database"

	"github.com/G-Research/yunikorn-history-server/test/config"
)

func TestClient_sync_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	_, repo, cleanupDB := setupDatabase(t, ctx)
	t.Cleanup(cleanupDB)
	eventRepository := repository.NewInMemoryEventRepository()

	c := NewRESTClient(config.GetTestYunikornConfig())
	s := NewService(repo, eventRepository, c)

	go func() { _ = s.Run(ctx) }()

	assert.Eventually(t, func() bool {
		return s.workqueue.Started()
	}, 500*time.Millisecond, 50*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	err := s.sync(ctx)
	if err != nil {
		t.Errorf("error starting up client: %v", err)
	}

	assert.Eventually(t, func() bool {
		partitions, err := repo.GetAllPartitions(ctx)
		if err != nil {
			t.Logf("error getting partitions: %v", err)
		}
		return len(partitions) > 0
	}, 10*time.Second, 250*time.Millisecond)
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

	now := time.Now().Unix()
	tests := []struct {
		name          string
		setup         func() *httptest.Server
		partitions    []*dao.PartitionInfo
		existingQueue *dao.PartitionQueueDAOInfo
		expected      []*model.PartitionQueueDAOInfo
		wantErr       bool
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
			partitions: []*dao.PartitionInfo{
				{Name: "default"},
			},
			existingQueue: nil,
			expected: []*model.PartitionQueueDAOInfo{
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
			partitions: []*dao.PartitionInfo{
				{Name: "default"},
			},
			existingQueue: &dao.PartitionQueueDAOInfo{
				QueueName: "root",
				Partition: "default",
				Children: []dao.PartitionQueueDAOInfo{
					{
						QueueName: "root.child-1",
					},
				},
			},
			expected: []*model.PartitionQueueDAOInfo{
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
			partitions: []*dao.PartitionInfo{
				{Name: "default"},
			},
			existingQueue: &dao.PartitionQueueDAOInfo{
				QueueName: "root",
				Partition: "default",
				Children: []dao.PartitionQueueDAOInfo{
					{
						QueueName: "root.child-1",
					},
				},
			},
			expected: []*model.PartitionQueueDAOInfo{
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
				{
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root.child-1",
						Partition: "default",
					},
					DeletedAt: &now,
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
			partitions:    []*dao.PartitionInfo{{Name: "default"}},
			existingQueue: nil,
			expected:      nil,
			wantErr:       true,
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
			partitions: []*dao.PartitionInfo{
				{Name: "default"},
				{Name: "secondary"},
				{Name: "third"},
			},
			existingQueue: nil,
			expected: []*model.PartitionQueueDAOInfo{
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
			partitions:    []*dao.PartitionInfo{{Name: "default"}},
			existingQueue: nil,
			expected: []*model.PartitionQueueDAOInfo{
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
			// seed the existing queues
			if tt.existingQueue != nil {
				if err := repo.AddQueues(ctx, nil, []*dao.PartitionQueueDAOInfo{tt.existingQueue}); err != nil {
					t.Fatalf("could not seed queue: %v", err)
				}
			}

			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))
			s := NewService(repo, eventRepository, client)

			// Create a cancellable context for this specific service
			ctx, cancel := context.WithCancel(context.Background())

			// Start the service in a goroutine
			go func() {
				_ = s.Run(ctx)
			}()

			// Ensure workqueue is started
			assert.Eventually(t, func() bool {
				return s.workqueue.Started()
			}, 500*time.Millisecond, 50*time.Millisecond)
			time.Sleep(100 * time.Millisecond)

			// Cleanup after each test case
			t.Cleanup(func() {
				cancel()
				s.workqueue.Shutdown()
			})

			_, err := s.syncQueues(context.Background(), tt.partitions)
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
	t.Cleanup(cancel)

	pool, repo, cleanupDB := setupDatabase(t, ctx)
	t.Cleanup(cleanupDB)
	eventRepository := repository.NewInMemoryEventRepository()

	tests := []struct {
		name               string
		setup              func() *httptest.Server
		existingPartitions []*dao.PartitionInfo
		expected           []*model.PartitionInfo
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
			expected: []*model.PartitionInfo{
				{PartitionInfo: dao.PartitionInfo{Name: "default"}},
				{PartitionInfo: dao.PartitionInfo{Name: "secondary"}},
			},
			wantErr: false,
		},
		{
			name: "Should mark secondary partition as deleted in DB",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.PartitionInfo{
						{Name: "default"},
					}
					writeResponse(t, w, response)
				}))
			},
			existingPartitions: []*dao.PartitionInfo{
				{Name: "default"},
				{Name: "secondary"},
			},
			expected: []*model.PartitionInfo{
				{PartitionInfo: dao.PartitionInfo{Name: "default"}},
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
			// seed the existing partitions
			if tt.existingPartitions != nil {
				if err := repo.UpsertPartitions(ctx, tt.existingPartitions); err != nil {
					t.Fatalf("could not seed partition: %v", err)
				}
			}

			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))
			s := NewService(repo, eventRepository, client)

			// Start the service
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				_ = s.Run(ctx)
			}()

			// Ensure workqueue is started
			assert.Eventually(t, func() bool {
				return s.workqueue.Started()
			}, 500*time.Millisecond, 50*time.Millisecond)
			time.Sleep(100 * time.Millisecond)

			// Cleanup after each test case
			t.Cleanup(func() {
				cancel()
				s.workqueue.Shutdown()
			})

			_, err := s.syncPartitions(ctx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var partitionsInDB []*model.PartitionInfo
			assert.Eventually(t, func() bool {
				partitionsInDB, err = s.repo.GetActivePartitions(ctx)
				if err != nil {
					t.Logf("error getting partitions: %v", err)
				}
				return len(partitionsInDB) == len(tt.expected)
			}, 5*time.Second, 50*time.Millisecond)

			for _, target := range tt.expected {
				if !isPartitionPresent(partitionsInDB, target) {
					t.Errorf("Partition %s is not found in the DB", target.Name)
				}
			}
		})
	}
}

func isQueuePresent(queuesInDB []*model.PartitionQueueDAOInfo, targetQueue *model.PartitionQueueDAOInfo) bool {
	for _, dbQueue := range queuesInDB {
		if dbQueue.QueueName == targetQueue.QueueName && dbQueue.Partition == targetQueue.Partition {
			// Check if DeletedAt fields are either both nil or both non-nil
			if (dbQueue.DeletedAt == nil && targetQueue.DeletedAt != nil) || (dbQueue.DeletedAt != nil && targetQueue.DeletedAt == nil) {
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

func isPartitionPresent(partitionsInDB []*model.PartitionInfo, targetPartition *model.PartitionInfo) bool {
	for _, dbPartition := range partitionsInDB {
		if dbPartition.Name == targetPartition.Name {
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
		database.DropTestSchema(ctx, t, schema)
	}

	return pool, repo, cleanup
}
