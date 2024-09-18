package yunikorn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

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

	repo, cleanupDB := setupDatabase(t, ctx)
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

	repo, cleanupDB := setupDatabase(t, ctx)
	t.Cleanup(cleanupDB)
	eventRepository := repository.NewInMemoryEventRepository()

	tests := []struct {
		name           string
		setup          func() *httptest.Server
		partitions     []*dao.PartitionInfo
		existingQueues []*dao.PartitionQueueDAOInfo
		expected       []*dao.PartitionQueueDAOInfo
		wantErr        bool
	}{
		{
			name: "Sync queues with no existing queues",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
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
				{
					Name: "default",
				},
			},
			existingQueues: nil,
			expected: []*dao.PartitionQueueDAOInfo{
				{QueueName: "root"},
				{QueueName: "root.child-1"},
				{QueueName: "root.child-1.1"},
				{QueueName: "root.child-1.2"},
			},
			wantErr: false,
		},
		{
			name: "Sync queues with existing queues in DB",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
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
				{
					Name: "default",
				},
			},
			existingQueues: []*dao.PartitionQueueDAOInfo{
				{QueueName: "root"},
				{QueueName: "root.child-1"},
			},
			expected: []*dao.PartitionQueueDAOInfo{
				{QueueName: "root"},
				{QueueName: "root.child-2"},
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
			partitions:     []*dao.PartitionInfo{{Name: "default"}},
			existingQueues: nil,
			expected:       nil,
			wantErr:        true,
		},
		{
			name: "Sync queues with multiple partitions",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
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
			},
			existingQueues: nil,
			expected: []*dao.PartitionQueueDAOInfo{
				{QueueName: "root"},
				{QueueName: "root.child-1"},
				{QueueName: "root.child-2"},
				{QueueName: "root"},
				{QueueName: "root.child-1"},
				{QueueName: "root.child-2"},
			},
			wantErr: false,
		},
		{
			name: "Sync queues with deeply nested queues",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := dao.PartitionQueueDAOInfo{
						QueueName: "root",
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
			partitions:     []*dao.PartitionInfo{{Name: "default"}},
			existingQueues: nil,
			expected: []*dao.PartitionQueueDAOInfo{
				{QueueName: "root"},
				{QueueName: "root.child-1"},
				{QueueName: "root.child-1.1"},
				{QueueName: "root.child-1.1.1"},
				{QueueName: "root.child-1.1.2"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			queues, err := s.syncQueues(context.Background(), tt.partitions)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if diff := cmp.Diff(tt.expected, queues, cmpopts.IgnoreFields(dao.PartitionQueueDAOInfo{}, "Children")); diff != "" {
				t.Errorf("Mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func setupDatabase(t *testing.T, ctx context.Context) (repository.Repository, func()) {
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

	return repo, cleanup
}
