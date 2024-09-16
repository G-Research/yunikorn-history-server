package webservice

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/model"
)

func TestWebServiceServeSPA(t *testing.T) {
	ws := &WebService{
		config: config.YHSConfig{
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
	tt := map[string]struct {
		queues  []*model.PartitionQueueDAOInfo
		want    *dao.PartitionQueueDAOInfo
		wantErr bool
	}{
		"no queues": {
			queues: []*model.PartitionQueueDAOInfo{},
			want:   nil,
		},
		"root queue": {
			queues: []*model.PartitionQueueDAOInfo{
				{
					Id: "1",
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
					},
				},
			},
			want: &dao.PartitionQueueDAOInfo{
				QueueName: "root",
			},
		},
		"two level tree": {
			queues: []*model.PartitionQueueDAOInfo{
				{
					Id: "2",
					ParentId: sql.NullString{
						String: "1",
						Valid:  true,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "child",
					},
				},
				{
					Id: "1",
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "root",
					},
				},
			},
			want: &dao.PartitionQueueDAOInfo{
				QueueName: "root",
				Children: []dao.PartitionQueueDAOInfo{
					{
						QueueName: "child",
					},
				},
			},
		},
		"no root queue": {
			queues: []*model.PartitionQueueDAOInfo{
				{
					Id: "2",
					ParentId: sql.NullString{
						String: "1",
						Valid:  true,
					},
					PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{
						QueueName: "child",
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got, err := buildPartitionQueueTree(context.TODO(), tc.queues)
			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, got)
		})
	}
}
