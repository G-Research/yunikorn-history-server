package model

import (
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
)

type ApplicationDAOInfo struct {
	CreatedAt time.Time `json:"createdAt"`
	dao.ApplicationDAOInfo
}
