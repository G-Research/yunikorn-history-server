package webservice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
