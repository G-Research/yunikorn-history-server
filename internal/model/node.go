package model

import (
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

type Node struct {
	ModelMetadata   `json:",inline"`
	Partition       *string `json:"partition"`
	dao.NodeDAOInfo `json:",inline"`
}

func (n *Node) MergeFrom(other *dao.NodeDAOInfo) {
	n.NodeDAOInfo = *other
}
