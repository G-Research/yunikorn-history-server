package model

import (
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
)

func TestNodeMergeFrom(t *testing.T) {
	meta := Metadata{
		CreatedAtNano: time.Now().UnixNano(),
	}

	tt := map[string]struct {
		node Node
		dao  *dao.NodeDAOInfo
		want Node
	}{
		"apply new allocations": {
			node: Node{
				Metadata: meta,
			},
			dao: &dao.NodeDAOInfo{
				NodeID: "1",
				Allocations: []*dao.AllocationDAOInfo{
					{
						AllocationKey: "alloc-1",
					},
				},
			},
			want: Node{
				Metadata: meta,
				NodeDAOInfo: dao.NodeDAOInfo{
					NodeID: "1",
					Allocations: []*dao.AllocationDAOInfo{
						{
							AllocationKey: "alloc-1",
						},
					},
				},
			},
		},
		"append new allocation": {
			node: Node{
				Metadata: meta,
				NodeDAOInfo: dao.NodeDAOInfo{
					Allocations: []*dao.AllocationDAOInfo{
						{
							AllocationKey: "alloc-1",
						},
					},
				},
			},
			dao: &dao.NodeDAOInfo{
				Allocations: []*dao.AllocationDAOInfo{
					{
						AllocationKey: "alloc-2",
					},
				},
			},
			want: Node{
				Metadata: meta,
				NodeDAOInfo: dao.NodeDAOInfo{
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
		"apply new reservations": {
			node: Node{
				Metadata: meta,
			},
			dao: &dao.NodeDAOInfo{
				Reservations: []string{"res-1"},
			},
			want: Node{
				Metadata: meta,
				NodeDAOInfo: dao.NodeDAOInfo{
					Reservations: []string{"res-1"},
				},
			},
		},
		"append new reservation": {
			node: Node{
				Metadata: meta,
				NodeDAOInfo: dao.NodeDAOInfo{
					Reservations: []string{"res-1"},
				},
			},
			dao: &dao.NodeDAOInfo{
				Reservations: []string{"res-2"},
			},
			want: Node{
				Metadata: meta,
				NodeDAOInfo: dao.NodeDAOInfo{
					Reservations: []string{"res-1", "res-2"},
				},
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			tc.node.MergeFrom(tc.dao)
			assert.Equal(t, tc.want, tc.node)
		})
	}
}
