package repository

import (
	"context"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"

	"github.com/G-Research/yunikorn-history-server/internal/model"
)

//go:generate mockgen -destination=mock_repository.go -package=repository github.com/G-Research/yunikorn-history-server/internal/database/repository Repository
type Repository interface {
	InsertApplication(ctx context.Context, app *model.Application) error
	UpdateApplication(ctx context.Context, app *model.Application) error
	GetLatestApplicationByApplicationID(ctx context.Context, appID string) (*model.Application, error)
	GetLatestApplicationsByApplicationID(ctx context.Context) ([]*model.Application, error)
	GetAllApplications(ctx context.Context, filters ApplicationFilters) ([]*model.Application, error)
	GetAppsPerPartitionPerQueue(ctx context.Context, partition, queue string, filters ApplicationFilters) ([]*model.Application, error)
	UpdateHistory(
		ctx context.Context,
		apps []*dao.ApplicationHistoryDAOInfo,
		containers []*dao.ContainerHistoryDAOInfo,
	) error
	GetApplicationsHistory(ctx context.Context, filters HistoryFilters) ([]*dao.ApplicationHistoryDAOInfo, error)
	GetContainersHistory(ctx context.Context, filters HistoryFilters) ([]*dao.ContainerHistoryDAOInfo, error)
	UpsertNodes(ctx context.Context, nodes []*dao.NodeDAOInfo, partition string) error
	InsertNodeUtilizations(ctx context.Context, partitionNodesUtil []*dao.PartitionNodesUtilDAOInfo) error
	GetNodesPerPartition(ctx context.Context, partition string, filters NodeFilters) ([]*dao.NodeDAOInfo, error)
	GetNodeUtilizations(ctx context.Context, filters NodeUtilFilters) ([]*dao.PartitionNodesUtilDAOInfo, error)
	UpsertPartitions(ctx context.Context, partitions []*dao.PartitionInfo) error
	GetAllPartitions(ctx context.Context, filters PartitionFilters) ([]*model.PartitionInfo, error)
	GetActivePartitions(ctx context.Context) ([]*model.PartitionInfo, error)
	DeleteInactivePartitions(ctx context.Context, activePartitions []*dao.PartitionInfo) error
	AddQueues(ctx context.Context, parentId *string, queues []*dao.PartitionQueueDAOInfo) error
	UpdateQueue(ctx context.Context, queue *dao.PartitionQueueDAOInfo) error
	UpsertQueues(ctx context.Context, queues []*dao.PartitionQueueDAOInfo) error
	GetAllQueues(ctx context.Context) ([]*model.PartitionQueueDAOInfo, error)
	GetQueuesPerPartition(ctx context.Context, partition string) ([]*model.PartitionQueueDAOInfo, error)
	GetQueue(ctx context.Context, partition, queueName string) (*model.PartitionQueueDAOInfo, error)
	DeleteQueues(ctx context.Context, queues []*model.PartitionQueueDAOInfo) error
}
