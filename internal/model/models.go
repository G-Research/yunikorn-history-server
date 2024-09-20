package model

import (
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
)

type ApplicationDAOInfo struct {
	CreatedAt              time.Time `json:"createdAt"`
	QueueID                string    `json:"queueId"`
	dao.ApplicationDAOInfo `json:",inline"`
}

type PartitionQueueDAOInfo struct {
	Id                        string  `json:"id"`
	ParentId                  *string `json:"parentId,omitempty"`
	dao.PartitionQueueDAOInfo `json:",inline"`
	Children                  []*PartitionQueueDAOInfo `json:"children,omitempty"`
	CreatedAt                 *int64                   `json:"createdAt,omitempty"`
	DeletedAt                 *int64                   `json:"deletedAt,omitempty"`
}
