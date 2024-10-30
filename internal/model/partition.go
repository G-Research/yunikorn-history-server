package model

import (
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

type Partition struct {
	Metadata          `json:",inline"`
	dao.PartitionInfo `json:",inline"`
}

func (p *Partition) MergeFrom(other *dao.PartitionInfo) {
	p.PartitionInfo = *other
}
