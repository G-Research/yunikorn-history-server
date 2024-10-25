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
	GetApplicationByID(ctx context.Context, id string) (*model.Application, error)
	DeleteApplicationsNotInIDs(ctx context.Context, ids []string, deletedAtNano int64) error
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
	GetNodeUtilizations(ctx context.Context, filters NodeUtilFilters) ([]*dao.PartitionNodesUtilDAOInfo, error)
	GetNodesPerPartition(ctx context.Context, partition string, filters NodeFilters) ([]*dao.NodeDAOInfo, error)
	InsertPartition(ctx context.Context, partition *model.Partition) error
	UpdatePartition(ctx context.Context, partition *model.Partition) error
	GetAllPartitions(ctx context.Context, filters PartitionFilters) ([]*model.Partition, error)
	GetLatestPartitionsGroupedByName(ctx context.Context) ([]*model.Partition, error)
	InsertQueue(ctx context.Context, q *model.Queue) error
	GetQueueInPartition(ctx context.Context, partition, queueName string) (*model.Queue, error)
	UpdateQueue(ctx context.Context, queue *model.Queue) error
	GetAllQueues(ctx context.Context) ([]*model.Queue, error)
	GetQueuesInPartition(ctx context.Context, partition string) ([]*model.Queue, error)
	DeleteQueuesNotInIDs(ctx context.Context, partition string, ids []string, deletedAtNano int64) error
}
