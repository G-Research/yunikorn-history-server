package model

import (
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

type ModelMetadata struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	DeletedAt *int64 `json:"deletedAt,omitempty"`
}

type PartitionQueueDAOInfo struct {
	Id                        string  `json:"id"`
	ParentId                  *string `json:"parentId,omitempty"`
	dao.PartitionQueueDAOInfo `json:",inline"`
	Children                  []*PartitionQueueDAOInfo `json:"children,omitempty"`
	CreatedAt                 *int64                   `json:"createdAt,omitempty"`
	DeletedAt                 *int64                   `json:"deletedAt,omitempty"`
}

type PartitionInfo struct {
	Id                string `json:"id"`
	dao.PartitionInfo `json:",inline"`
	DeletedAt         *int64 `json:"deletedAt,omitempty"`
}
