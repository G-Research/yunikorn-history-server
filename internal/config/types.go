package config

import (
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
)

type PostgresConfig struct {
	Host                string
	Port                int
	DbName              string
	Username            string
	Password            string
	PoolMaxConns        int
	PoolMinConns        int
	PoolMaxConnLifetime time.Duration
	PoolMaxConnIdleTime time.Duration
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

type NodesResponse struct {
	Nodes []*dao.NodeDAOInfo `json:"nodes"`
}
