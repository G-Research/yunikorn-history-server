package repository

import (
	"context"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/google/uuid"

	"github.com/G-Research/yunikorn-history-server/internal/model"
)

//go:generate mockgen -destination=mock_repository.go -package=repository github.com/G-Research/yunikorn-history-server/internal/database/repository Repository
type Repository interface {
	UpsertApplications(ctx context.Context, apps []*dao.ApplicationDAOInfo) error
	GetAllApplications(ctx context.Context, filters ApplicationFilters) ([]*model.ApplicationDAOInfo, error)
	GetAppsPerPartitionPerQueue(ctx context.Context, partition, queue string, filters ApplicationFilters) ([]*model.ApplicationDAOInfo, error)
	UpdateHistory(
		ctx context.Context,
		apps []*dao.ApplicationHistoryDAOInfo,
		containers []*dao.ContainerHistoryDAOInfo,
	) error
	GetApplicationsHistory(ctx context.Context) ([]*dao.ApplicationHistoryDAOInfo, error)
	GetContainersHistory(ctx context.Context) ([]*dao.ContainerHistoryDAOInfo, error)
	UpsertNodes(ctx context.Context, nodes []*dao.NodeDAOInfo, partition string) error
	InsertNodeUtilizations(ctx context.Context, uuid uuid.UUID, partitionNodesUtil []*dao.PartitionNodesUtilDAOInfo) error
	GetNodeUtilizations(ctx context.Context) ([]*dao.PartitionNodesUtilDAOInfo, error)
	GetNodesPerPartition(ctx context.Context, partition string) ([]*dao.NodeDAOInfo, error)
	UpsertPartitions(ctx context.Context, partitions []*dao.PartitionInfo) error
	GetAllPartitions(ctx context.Context) ([]*dao.PartitionInfo, error)
	UpsertQueues(ctx context.Context, queues []*dao.PartitionQueueDAOInfo) error
	GetAllQueues(ctx context.Context) ([]*dao.PartitionQueueDAOInfo, error)
	GetQueuesPerPartition(ctx context.Context, partition string) ([]*model.PartitionQueueDAOInfo, error)
}
