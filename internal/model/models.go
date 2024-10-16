package model

import (
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

type ModelMetadata struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	DeletedAt *int64 `json:"deletedAt,omitempty"`
}

type ApplicationDAOInfo struct {
	ID                     string    `json:"id"`
	CreatedAt              time.Time `json:"createdAt"`
	QueueID                string    `json:"queueId"`
	dao.ApplicationDAOInfo `json:",inline"`
}
