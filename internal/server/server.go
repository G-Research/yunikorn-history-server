package server

import (
	"context"
	"github.com/G-Research/yunikorn-history-server/internal/oas"
	"github.com/G-Research/yunikorn-history-server/internal/repository"
)

type Server struct {
	storage *repository.RepoPostgres
}

func NewServer(storage *repository.RepoPostgres) *Server {
	return &Server{
		storage: storage,
	}
}

func (s Server) GetAppsHistory(ctx context.Context) (oas.GetAppsHistoryRes, error) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetAppsPerPartitionPerQueue(ctx context.Context, params oas.GetAppsPerPartitionPerQueueParams) (oas.GetAppsPerPartitionPerQueueRes, error) {
	//TODO implement me
	panic("implement me")
	gsdljgdsnl
	gljdngdfl

}

func (s Server) GetContainersHistory(ctx context.Context) (oas.GetContainersHistoryRes, error) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetEventStatistics(ctx context.Context) (oas.GetEventStatisticsRes, error) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetNodeUtilizations(ctx context.Context) (oas.GetNodeUtilizationsRes, error) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetNodesPerPartition(ctx context.Context, params oas.GetNodesPerPartitionParams) (oas.GetNodesPerPartitionRes, error) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetPartitions(ctx context.Context) (oas.GetPartitionsRes, error) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetQueuesPerPartition(ctx context.Context, params oas.GetQueuesPerPartitionParams) (oas.GetQueuesPerPartitionRes, error) {
	//TODO implement me
	panic("implement me")
}

var _ oas.Handler = &Server{}
