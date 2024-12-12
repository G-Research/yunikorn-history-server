package webservice

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/G-Research/unicorn-history-server/internal/database/repository"
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/G-Research/unicorn-history-server/internal/config"
	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/unicorn-history-server/internal/util"
)

func TestWebServiceServeSPA(t *testing.T) {
	ws := &WebService{
		config: config.UHSConfig{
			AssetsDir: "testdir",
		},
	}

	tt := map[string]struct {
		path            string
		wantFile        string
		wantContentType string
	}{
		"root path": {
			path:            "/",
			wantFile:        "testdir/index.html",
			wantContentType: "text/html; charset=utf-8",
		},
		"js file": {
			path:            "/js/app.js",
			wantFile:        "testdir/js/app.js",
			wantContentType: "text/javascript",
		},
		"file not found": {
			path:            "/notfound",
			wantFile:        "testdir/index.html",
			wantContentType: "text/html; charset=utf-8",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			ws.serveSPA(rec, req)

			f, err := os.Open(tc.wantFile)
			require.NoError(t, err)
			defer f.Close()

			wantContent, err := os.ReadFile(tc.wantFile)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, string(wantContent), rec.Body.String())
			assert.Contains(t, rec.Header().Get("Content-Type"), tc.wantContentType)
		})
	}
}

func TestBuildPartitionQueueTree(t *testing.T) {
	now := time.Now().UnixNano()
	tt := map[string]struct {
		queues  []*model.Queue
		want    []*model.Queue
		wantErr bool
	}{
		"no queues": {
			queues: []*model.Queue{},
			want:   nil,
		},
		"root queue": {
			queues: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "1",
						QueueName: "root",
					},
				},
			},
			want: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "1",
						QueueName: "root",
					},
				},
			},
		},
		"no root queue": {
			queues: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "2",
						QueueName: "child",
						ParentID:  util.ToPtr("1"),
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
		"multiple root queues": {
			queues: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "2",
						QueueName: "child-1",
						ParentID:  util.ToPtr("1"),
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root-1",
						ID:        "1",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "22",
						QueueName: "child-2",
						ParentID:  util.ToPtr("11"),
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "11",
						QueueName: "root-2",
					},
				},
			},
			want: []*model.Queue{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "1",
						QueueName: "root-1",
					},
					Children: []*model.Queue{
						{
							Metadata: model.Metadata{
								CreatedAtNano: now,
							},
							PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
								ID:        "2",
								QueueName: "child-1",
								ParentID:  util.ToPtr("1"),
							},
						},
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						ID:        "11",
						QueueName: "root-2",
					},
					Children: []*model.Queue{
						{
							Metadata: model.Metadata{
								CreatedAtNano: now,
							},
							PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
								ID:        "22",
								QueueName: "child-2",
								ParentID:  util.ToPtr("11"),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got, err := buildPartitionQueueTrees(context.TODO(), tc.queues)
			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetAppsPerPartitionPerQueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)

	tests := []struct {
		name           string
		partitionID    string
		queueID        string
		expectedApps   []*model.Application
		expectedStatus int
	}{
		{
			name:        "Apps found",
			partitionID: "1",
			queueID:     "1",
			expectedApps: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: time.Now().UnixNano(),
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "1",
						ApplicationID: "app1",
						Partition:     "default",
						PartitionID:   "1",
						QueueID:       util.ToPtr("1"),
						QueueName:     "root.default",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No apps found",
			partitionID:    "1",
			queueID:        "1",
			expectedApps:   nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetAppsPerPartitionPerQueue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.expectedApps, nil)

			ws := &WebService{repository: mockRepo}

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/partitions/%s/queues/%s/apps", tt.partitionID, tt.queueID), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			ws.getAppsPerPartitionPerQueue(restful.NewRequest(req), restful.NewResponse(rr))
			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetPartitions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)

	tests := []struct {
		name               string
		expectedPartitions []*model.Partition
		expectedStatus     int
	}{
		{
			name: "Partitions found",
			expectedPartitions: []*model.Partition{
				{
					Metadata: model.Metadata{
						CreatedAtNano: time.Now().UnixNano(),
					},
					PartitionInfo: dao.PartitionInfo{
						ID:                      "1",
						Name:                    "default",
						ClusterID:               "cluster1",
						State:                   "Active",
						LastStateTransitionTime: time.Now().Add(-1 * time.Hour).UnixMilli(),
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "No Partition found",
			expectedPartitions: nil,
			expectedStatus:     http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetAllPartitions(gomock.Any(), gomock.Any()).
				Return(tt.expectedPartitions, nil)

			ws := &WebService{repository: mockRepo}

			req, err := http.NewRequest(http.MethodGet, "/api/v1/partitions", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			ws.getPartitions(restful.NewRequest(req), restful.NewResponse(rr))
			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetQueuesPerPartition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)

	tests := []struct {
		name           string
		partitionID    string
		expectedQueues []*model.Queue
		expectedStatus int
	}{
		{
			name:        "Apps found",
			partitionID: "1",
			expectedQueues: []*model.Queue{
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
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No queue found",
			partitionID:    "1",
			expectedQueues: nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetQueuesInPartition(gomock.Any(), gomock.Any()).
				Return(tt.expectedQueues, nil)

			ws := &WebService{repository: mockRepo}

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/partitions/%s/queues", tt.partitionID), nil)

			require.NoError(t, err)

			rr := httptest.NewRecorder()

			ws.getQueuesPerPartition(restful.NewRequest(req), restful.NewResponse(rr))
			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetNodesPerPartition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)

	tests := []struct {
		name           string
		partitionID    string
		expectedNodes  []*model.Node
		expectedStatus int
	}{
		{
			name:        "Nodes found",
			partitionID: "1",
			expectedNodes: []*model.Node{
				{
					Metadata: model.Metadata{
						CreatedAtNano: time.Now().UnixNano(),
					},
					NodeDAOInfo: dao.NodeDAOInfo{
						ID:          "1",
						PartitionID: "1",
						NodeID:      "node1",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No nodes found",
			partitionID:    "1",
			expectedNodes:  nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetNodesPerPartition(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.expectedNodes, nil)

			ws := &WebService{repository: mockRepo}

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/partitions/%s/nodes", tt.partitionID), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			ws.getNodesPerPartition(restful.NewRequest(req), restful.NewResponse(rr))
			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetAppsHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)

	tests := []struct {
		name           string
		expectedApps   []*model.AppHistory
		expectedStatus int
	}{
		{
			name: "App History found",
			expectedApps: []*model.AppHistory{
				{
					Metadata: model.Metadata{
						CreatedAtNano: time.Now().UnixNano(),
					},
					ApplicationHistoryDAOInfo: dao.ApplicationHistoryDAOInfo{
						Timestamp:         time.Now().UnixNano(),
						TotalApplications: "5",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No apps found",
			expectedApps:   nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetApplicationsHistory(gomock.Any(), gomock.Any()).
				Return(tt.expectedApps, nil)

			ws := &WebService{repository: mockRepo}

			req, err := http.NewRequest(http.MethodGet, "/api/v1/apps/history", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			ws.getAppsHistory(restful.NewRequest(req), restful.NewResponse(rr))
			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetContainersHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)

	tests := []struct {
		name           string
		expectedApps   []*model.ContainerHistory
		expectedStatus int
	}{
		{
			name: "Container History found",
			expectedApps: []*model.ContainerHistory{
				{
					Metadata: model.Metadata{
						CreatedAtNano: time.Now().UnixNano(),
					},
					ContainerHistoryDAOInfo: dao.ContainerHistoryDAOInfo{
						Timestamp:       time.Now().UnixNano(),
						TotalContainers: "5",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No containers found",
			expectedApps:   nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetContainersHistory(gomock.Any(), gomock.Any()).
				Return(tt.expectedApps, nil)

			ws := &WebService{repository: mockRepo}

			req, err := http.NewRequest(http.MethodGet, "/api/v1/containers/history", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			ws.getContainersHistory(restful.NewRequest(req), restful.NewResponse(rr))
			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
