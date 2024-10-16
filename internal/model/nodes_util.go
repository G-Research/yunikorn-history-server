package model

import (
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

type NodesUtil struct {
	ModelMetadata                 `json:",inline"`
	dao.PartitionNodesUtilDAOInfo `json:",inline"`
}

func (nu *NodesUtil) MergeFrom(other *dao.PartitionNodesUtilDAOInfo) {
	nu.PartitionNodesUtilDAOInfo = *other
}
