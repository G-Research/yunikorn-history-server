package model

import (
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

type AppHistory struct {
	ModelMetadata                 `json:",inline"`
	dao.ApplicationHistoryDAOInfo `json:",inline"`
}

func (h *AppHistory) MergeFromAppHistory(other *dao.ApplicationHistoryDAOInfo) {
	h.ApplicationHistoryDAOInfo = *other
}

type ContainerHistory struct {
	ModelMetadata               `json:",inline"`
	dao.ContainerHistoryDAOInfo `json:",inline"`
}

func (h *ContainerHistory) MergeFromContainerHistory(other *dao.ContainerHistoryDAOInfo) {
	h.ContainerHistoryDAOInfo = *other
}
