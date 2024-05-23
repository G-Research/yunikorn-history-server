package config

import (
	"fmt"
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	EventCounts = contextKey("eventCounts")
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

type EventTypeKey struct {
	Type       si.EventRecord_Type
	ChangeType si.EventRecord_ChangeType
}

func (ek EventTypeKey) MarshalText() (text []byte, err error) {
	str := fmt.Sprintf("%s.%s", ek.Type, ek.ChangeType)
	return []byte(str), nil
}

type EventTypeCounts map[EventTypeKey]int
