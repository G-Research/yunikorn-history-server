package model

import (
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
)

type ModelMetadata struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	DeletedAt *int64 `json:"deletedAt,omitempty"`
}

type Application struct {
	ModelMetadata          `json:",inline"`
	dao.ApplicationDAOInfo `json:",inline"`
}

func (app *Application) MergeFrom(appInfo dao.ApplicationDAOInfo) {
	oldAppInfo := app.ApplicationDAOInfo
	app.ApplicationDAOInfo = appInfo
	if len(app.Allocations) < len(oldAppInfo.Allocations) {
		app.Allocations = oldAppInfo.Allocations
	}
	if len(app.Requests) < len(oldAppInfo.Requests) {
		app.Requests = oldAppInfo.Requests
	}
	if len(app.StateLog) < len(oldAppInfo.StateLog) {
		app.StateLog = oldAppInfo.StateLog
	}
}

type ApplicationDAOInfo struct {
	ID                     string    `json:"id"`
	CreatedAt              time.Time `json:"createdAt"`
	QueueID                string    `json:"queueId"`
	dao.ApplicationDAOInfo `json:",inline"`
}

type PartitionQueueDAOInfo struct {
	ID                        string  `json:"id"`
	ParentID                  *string `json:"parentId,omitempty"`
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
