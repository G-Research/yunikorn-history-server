package config

import (
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
)

type PostgresConfig struct {
	PoolMaxOpenConns    int
	PoolMaxIdleConns    int
	PoolMaxConnLifetime time.Duration
	Connection          map[string]string
}

type ECConfig struct {
	// Other settings for event-collector go here

	PostgresConfig PostgresConfig
}

type PartitionsResponse struct {
	Partitions []*dao.PartitionInfo `json:"partitions"`
}

type QueuesResponse struct {
	Queues []*dao.PartitionQueueDAOInfo `json:"queues"`
}

type AppsResponse struct {
	Apps []*dao.ApplicationDAOInfo `json:"apps"`
}
