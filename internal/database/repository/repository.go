package repository

import (
	"context"

	"github.com/G-Research/unicorn-history-server/internal/model"
)

//go:generate mockgen -destination=mock_repository.go -package=repository github.com/G-Research/unicorn-history-server/internal/database/repository Repository
type Repository interface {
	InsertApplication(ctx context.Context, app *model.Application) error
	UpdateApplication(ctx context.Context, app *model.Application) error
	GetApplicationByID(ctx context.Context, id string) (*model.Application, error)
	DeleteApplicationsNotInIDs(ctx context.Context, ids []string, deletedAtNano int64) error
	GetAllApplications(ctx context.Context, filters ApplicationFilters) ([]*model.Application, error)
	GetAppsPerPartitionPerQueue(ctx context.Context, partitionID, queueID string, filters ApplicationFilters) ([]*model.Application, error)
	InsertAppHistory(ctx context.Context, appHistory *model.AppHistory) error
	InsertContainerHistory(ctx context.Context, containerHistory *model.ContainerHistory) error
	GetApplicationsHistory(ctx context.Context, filters HistoryFilters) ([]*model.AppHistory, error)
	GetContainersHistory(ctx context.Context, filters HistoryFilters) ([]*model.ContainerHistory, error)
	InsertNode(ctx context.Context, node *model.Node) error
	UpdateNode(ctx context.Context, node *model.Node) error
	GetNodeByID(ctx context.Context, id string) (*model.Node, error)
	DeleteNodesNotInIDs(ctx context.Context, ids []string, deletedAtNano int64) error
	GetNodesPerPartition(ctx context.Context, partition string, filters NodeFilters) ([]*model.Node, error)
	InsertPartition(ctx context.Context, partition *model.Partition) error
	UpdatePartition(ctx context.Context, partition *model.Partition) error
	GetAllPartitions(ctx context.Context, filters PartitionFilters) ([]*model.Partition, error)
	GetPartitionByID(ctx context.Context, id string) (*model.Partition, error)
	DeletePartitionsNotInIDs(ctx context.Context, ids []string, deletedatNano int64) error
	InsertQueue(ctx context.Context, q *model.Queue) error
	GetQueue(ctx context.Context, queueID string) (*model.Queue, error)
	UpdateQueue(ctx context.Context, queue *model.Queue) error
	GetAllQueues(ctx context.Context) ([]*model.Queue, error)
	GetQueuesInPartition(ctx context.Context, partitionID string) ([]*model.Queue, error)
	DeleteQueuesNotInIDs(ctx context.Context, ids []string, deletedAtNano int64) error
}
