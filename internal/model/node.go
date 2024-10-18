package model

import (
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

type Node struct {
	Metadata        `json:",inline"`
	Partition       *string `json:"partition"`
	dao.NodeDAOInfo `json:",inline"`
}

func (n *Node) MergeFrom(nodeInfo *dao.NodeDAOInfo) {
	n.NodeID = nodeInfo.NodeID
	n.HostName = nodeInfo.HostName
	n.RackName = nodeInfo.RackName
	n.Attributes = nodeInfo.Attributes
	n.Capacity = nodeInfo.Capacity
	n.Allocated = nodeInfo.Allocated
	n.Occupied = nodeInfo.Occupied
	n.Available = nodeInfo.Available
	n.Utilized = nodeInfo.Utilized
	n.Allocations = nodeInfo.Allocations
	n.Schedulable = nodeInfo.Schedulable
	n.IsReserved = nodeInfo.IsReserved

	lookup := make(map[string]struct{})
	if len(nodeInfo.Allocations) > 0 {
		for _, alloc := range n.Allocations {
			lookup[alloc.AllocationKey] = struct{}{}
		}
		for _, alloc := range nodeInfo.Allocations {
			if _, ok := lookup[alloc.AllocationKey]; !ok {
				n.Allocations = append(n.Allocations, alloc)
			}
		}
		clear(lookup)
	}

	if len(nodeInfo.Reservations) > 0 {
		for _, res := range n.Reservations {
			lookup[res] = struct{}{}
		}
		for _, res := range nodeInfo.Reservations {
			if _, ok := lookup[res]; !ok {
				n.Reservations = append(n.Reservations, res)
			}
		}
		clear(lookup)
	}
}
