package yunikorn

import (
	"context"
	"net/http"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

// Client defines the interface for interacting with the Yunikorn REST API.
//
//go:generate mockgen -destination=mock_client.go -package=yunikorn github.com/G-Research/unicorn-history-server/internal/yunikorn Client
type Client interface {
	GetPartitions(ctx context.Context) ([]*dao.PartitionInfo, error)
	GetPartitionQueues(ctx context.Context, partitionName string) (*dao.PartitionQueueDAOInfo, error)
	GetPartitionQueue(ctx context.Context, partitionName, queueName string) (*dao.PartitionQueueDAOInfo, error)
	GetApplications(ctx context.Context, partitionName, queueName string) ([]*dao.ApplicationDAOInfo, error)
	GetApplication(ctx context.Context, partitionName, queueName, appID string) (*dao.ApplicationDAOInfo, error)
	GetPartitionNodes(ctx context.Context, partitionName string) ([]*dao.NodeDAOInfo, error)
	GetAppsHistory(ctx context.Context) ([]*dao.ApplicationHistoryDAOInfo, error)
	GetContainersHistory(ctx context.Context) ([]*dao.ContainerHistoryDAOInfo, error)
	GetEventStream(ctx context.Context) (*http.Response, error)
	Healthcheck(ctx context.Context) (*dao.SchedulerHealthDAOInfo, error)
}
