package yunikorn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/oklog/ulid/v2"

	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/unicorn-history-server/internal/util"
)

// syncPartitions fetches partitions from the Yunikorn API and syncs them into the database
func (s *Service) syncPartitions(ctx context.Context, partitions []*dao.PartitionInfo) error {
	ids := make([]string, 0, len(partitions))
	for _, p := range partitions {
		ids = append(ids, p.ID)
	}

	now := time.Now().UnixNano()
	if err := s.repo.DeletePartitionsNotInIDs(ctx, ids, now); err != nil {
		return fmt.Errorf("could not delete partitions not in IDs: %w", err)
	}

	for _, p := range partitions {
		current, err := s.repo.GetPartitionByID(ctx, p.ID)
		fmt.Printf("Getting partition resulted in current: %+v, err: %v\n", current, err)
		if err != nil {
			fmt.Printf("Error getting partition: %v\n", err)
			partition := &model.Partition{
				Metadata: model.Metadata{
					CreatedAtNano: now,
				},
				PartitionInfo: *p,
			}

			if err := s.repo.InsertPartition(ctx, partition); err != nil {
				return fmt.Errorf("could not insert partition: %w", err)
			}
			continue
		}

		current.MergeFrom(p)

		if err := s.repo.UpdatePartition(ctx, current); err != nil {
			return fmt.Errorf("could not update partition: %w", err)
		}
	}
	return nil
}

func (s *Service) syncQueues(ctx context.Context, clientQueues []dao.PartitionQueueDAOInfo) error {
	var errs []error

	queues := flattenQueues(util.ToPtrSlice(clientQueues))

	ids := make([]string, 0, len(queues))
	for _, q := range queues {
		ids = append(ids, q.ID)
	}

	now := time.Now().UnixNano()
	if err := s.repo.DeleteQueuesNotInIDs(ctx, ids, now); err != nil {
		return fmt.Errorf("could not delete queues not in IDs: %w", err)
	}

	for _, q := range queues {
		current, err := s.repo.GetQueue(ctx, q.ID)
		if err != nil {
			queue := &model.Queue{
				Metadata: model.Metadata{
					CreatedAtNano: now,
				},
				PartitionQueueDAOInfo: *q,
			}
			if err := s.repo.InsertQueue(ctx, queue); err != nil {
				errs = append(errs, fmt.Errorf("could not insert queue: %w", err))
			}
			continue
		}

		current.MergeFrom(q)
		if err := s.repo.UpdateQueue(ctx, current); err != nil {
			errs = append(errs, fmt.Errorf("could not update queue: %w", err))
		}
	}
	return errors.Join(errs...)
}

// flattenQueues returns a list of all queues in the hierarchy in a flat array.
// Usually the returned queues are a single hierarchical structure, the root queue,
// and all other queues are children queues.
func flattenQueues(qs []*dao.PartitionQueueDAOInfo) []*dao.PartitionQueueDAOInfo {
	var queues []*dao.PartitionQueueDAOInfo
	for _, q := range qs {
		queues = append(queues, q)
		if len(q.Children) > 0 {
			queues = append(queues, flattenQueues(util.ToPtrSlice(q.Children))...)
		}
	}
	return queues
}

func (s *Service) syncNodes(ctx context.Context, daoNodes []*dao.NodesDAOInfo) error {
	var errs []error
	for _, nodesInfo := range daoNodes {
		nodes := nodesInfo.Nodes
		ids := make([]string, 0, len(nodes))
		for _, n := range nodes {
			ids = append(ids, n.ID)
		}
		nowNano := time.Now().UnixNano()
		if err := s.repo.DeleteNodesNotInIDs(ctx, ids, nowNano); err != nil {
			errs = append(errs, err)
		}

		for _, n := range nodes {
			current, err := s.repo.GetNodeByID(ctx, n.ID)
			if err != nil { // node not found
				node := &model.Node{
					Metadata: model.Metadata{
						CreatedAtNano: nowNano,
					},
					NodeDAOInfo: *n,
				}
				if err := s.repo.InsertNode(ctx, node); err != nil {
					errs = append(errs, fmt.Errorf("could not insert node %s: %v", n.NodeID, err))
				}
				continue
			}

			current.MergeFrom(n)
			if err := s.repo.UpdateNode(ctx, current); err != nil {
				errs = append(errs, fmt.Errorf("could not update node %s: %v", n.NodeID, err))
			}
		}
	}

	return errors.Join(errs...)
}

// syncApplications fetches applications for each queue and upserts them into the database
func (s *Service) syncApplications(ctx context.Context, applications []*dao.ApplicationDAOInfo) error {

	ids := make([]string, 0, len(applications))
	for _, a := range applications {
		ids = append(ids, a.ID)
	}

	nowNano := time.Now().UnixNano()
	if err := s.repo.DeleteApplicationsNotInIDs(ctx, ids, nowNano); err != nil {
		return err
	}

	for _, app := range applications {
		current, err := s.repo.GetApplicationByID(ctx, app.ID)
		if err != nil { // todo: Further check if error is not found
			application := &model.Application{
				Metadata: model.Metadata{
					CreatedAtNano: nowNano,
				},
				ApplicationDAOInfo: *app,
			}
			if err := s.repo.InsertApplication(ctx, application); err != nil {
				return err
			}
			continue
		}

		current.MergeFrom(app)
		if err := s.repo.UpdateApplication(ctx, current); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) syncAppHistory(ctx context.Context, appsHistory []*dao.ApplicationHistoryDAOInfo) error {
	var errs []error
	nowNano := time.Now().UnixNano()
	for _, ah := range appsHistory {
		history := &model.AppHistory{
			Metadata: model.Metadata{
				CreatedAtNano: nowNano,
			},
			ID:                        ulid.Make().String(),
			ApplicationHistoryDAOInfo: *ah,
		}
		if err := s.repo.InsertAppHistory(ctx, history); err != nil {
			errs = append(errs, fmt.Errorf("could not insert app history: %v", err))
		}
	}
	return errors.Join(errs...)
}

func (s *Service) syncContainerHistory(ctx context.Context, containersHistory []*dao.ContainerHistoryDAOInfo) error {
	var errs []error
	nowNano := time.Now().UnixNano()
	for _, ch := range containersHistory {
		history := &model.ContainerHistory{
			Metadata: model.Metadata{
				CreatedAtNano: nowNano,
			},
			ID:                      ulid.Make().String(),
			ContainerHistoryDAOInfo: *ch,
		}
		if err := s.repo.InsertContainerHistory(ctx, history); err != nil {
			errs = append(errs, fmt.Errorf("could not insert container history: %v", err))
		}
	}
	return errors.Join(errs...)
}
