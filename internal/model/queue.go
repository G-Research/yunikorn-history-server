package model

import "github.com/G-Research/yunikorn-core/pkg/webservice/dao"

type Queue struct {
	Metadata `json:",inline"`
	ParentID *string `json:"parentId"`
	// This field should be used instead of the dao.Children
	Children                  []*Queue `json:"children,omitempty"`
	dao.PartitionQueueDAOInfo `json:",inline"`
}

func (q *Queue) MergeFrom(qInfo *dao.PartitionQueueDAOInfo) {
	q.PartitionQueueDAOInfo = *qInfo
}
