package model

import (
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
)

func TestApplicationMergeFrom(t *testing.T) {
	meta := Metadata{
		CreatedAtNano: time.Now().UnixNano(),
	}

	tt := map[string]struct {
		app  Application
		dao  *dao.ApplicationDAOInfo
		want Application
	}{
		"apply new allocations": {
			app: Application{
				Metadata: meta,
			},
			dao: &dao.ApplicationDAOInfo{
				ID: "1",
				Allocations: []*dao.AllocationDAOInfo{
					{
						AllocationKey: "alloc-1",
					},
				},
			},
			want: Application{
				Metadata: meta,
				ApplicationDAOInfo: dao.ApplicationDAOInfo{
					ID: "1",
					Allocations: []*dao.AllocationDAOInfo{
						{
							AllocationKey: "alloc-1",
						},
					},
				},
			},
		},
		"append new allocation": {
			app: Application{
				Metadata: meta,
				ApplicationDAOInfo: dao.ApplicationDAOInfo{
					Allocations: []*dao.AllocationDAOInfo{
						{
							AllocationKey: "alloc-1",
						},
					},
				},
			},
			dao: &dao.ApplicationDAOInfo{
				Allocations: []*dao.AllocationDAOInfo{
					{
						AllocationKey: "alloc-2",
					},
				},
			},
			want: Application{
				Metadata: meta,
				ApplicationDAOInfo: dao.ApplicationDAOInfo{
					Allocations: []*dao.AllocationDAOInfo{
						{
							AllocationKey: "alloc-1",
						},
						{
							AllocationKey: "alloc-2",
						},
					},
				},
			},
		},
		"apply new requests": {
			app: Application{
				Metadata: meta,
			},
			dao: &dao.ApplicationDAOInfo{
				Requests: []*dao.AllocationAskDAOInfo{
					{
						AllocationKey: "ask-1",
					},
				},
			},
			want: Application{
				Metadata: meta,
				ApplicationDAOInfo: dao.ApplicationDAOInfo{
					Requests: []*dao.AllocationAskDAOInfo{
						{
							AllocationKey: "ask-1",
						},
					},
				},
			},
		},
		"append new request": {
			app: Application{
				Metadata: meta,
				ApplicationDAOInfo: dao.ApplicationDAOInfo{
					Requests: []*dao.AllocationAskDAOInfo{
						{
							AllocationKey: "ask-1",
						},
					},
				},
			},
			dao: &dao.ApplicationDAOInfo{
				Requests: []*dao.AllocationAskDAOInfo{
					{
						AllocationKey: "ask-2",
					},
				},
			},
			want: Application{
				Metadata: meta,
				ApplicationDAOInfo: dao.ApplicationDAOInfo{
					Requests: []*dao.AllocationAskDAOInfo{
						{
							AllocationKey: "ask-1",
						},
						{
							AllocationKey: "ask-2",
						},
					},
				},
			},
		},
		"append new request and allocation": {
			app: Application{
				Metadata: meta,
				ApplicationDAOInfo: dao.ApplicationDAOInfo{
					Allocations: []*dao.AllocationDAOInfo{
						{
							AllocationKey: "alloc-1",
						},
					},
					Requests: []*dao.AllocationAskDAOInfo{
						{
							AllocationKey: "ask-1",
						},
					},
				},
			},
			dao: &dao.ApplicationDAOInfo{
				Allocations: []*dao.AllocationDAOInfo{
					{
						AllocationKey: "alloc-2",
					},
				},
				Requests: []*dao.AllocationAskDAOInfo{
					{
						AllocationKey: "ask-2",
					},
				},
			},
			want: Application{
				Metadata: meta,
				ApplicationDAOInfo: dao.ApplicationDAOInfo{
					Allocations: []*dao.AllocationDAOInfo{
						{
							AllocationKey: "alloc-1",
						},
						{
							AllocationKey: "alloc-2",
						},
					},
					Requests: []*dao.AllocationAskDAOInfo{
						{
							AllocationKey: "ask-1",
						},
						{
							AllocationKey: "ask-2",
						},
					},
				},
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			tc.app.MergeFrom(tc.dao)
			assert.Equal(t, tc.want, tc.app)
		})
	}
}
