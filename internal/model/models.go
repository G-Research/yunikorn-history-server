package model

import (
	"database/sql"
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
)

type ApplicationDAOInfo struct {
	CreatedAt time.Time `json:"createdAt"`
	QueueID   string    `json:"queueId"`
	dao.ApplicationDAOInfo
}

type PartitionQueueDAOInfo struct {
	Id       string         `json:"id"`
	ParentId sql.NullString `json:"parentId,omitempty"`
	dao.PartitionQueueDAOInfo
	Children  []PartitionQueueDAOInfo `json:"children,omitempty"`
	CreatedAt sql.NullInt64           `json:"createdAt,omitempty"`
	DeletedAt sql.NullInt64           `json:"deletedAt,omitempty"`
}
