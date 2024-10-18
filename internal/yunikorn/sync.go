package yunikorn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"

	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/internal/workqueue"
)

// syncPartitions fetches partitions from the Yunikorn API and syncs them into the database
func (s *Service) syncPartitions(ctx context.Context) ([]*model.Partition, error) {
	// Get partitions from Yunikorn API and upsert into DB
	partitions, err := s.client.GetPartitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get partitions: %v", err)
	}

	ids := make([]string, 0, len(partitions))
	for _, p := range partitions {
		ids = append(ids, p.ID)
	}

	now := time.Now().UnixNano()
	if err := s.repo.DeletePartitionsNotInIDs(ctx, ids, now); err != nil {
		return nil, fmt.Errorf("could not delete partitions not in IDs: %w", err)
	}

	result := make([]*model.Partition, 0, len(partitions))
	for _, p := range partitions {
		current, err := s.repo.GetPartitionByID(ctx, p.ID)
		if err != nil {
			partition := &model.Partition{
				Metadata: model.Metadata{
					CreatedAtNano: now,
				},
				PartitionInfo: *p,
			}

			if err := s.repo.InsertPartition(ctx, partition); err != nil {
				return nil, fmt.Errorf("could not insert partition: %w", err)
			}

			result = append(result, partition)
			continue
		}

		current.MergeFrom(p)

		if err := s.repo.UpdatePartition(ctx, current); err != nil {
			return nil, fmt.Errorf("could not update partition: %w", err)
		}

		result = append(result, current)
	}

	return result, nil
}

// syncQueues fetches queues for each partition and upserts them into the database
func (s *Service) syncQueues(ctx context.Context, partitions []*model.Partition) error {
	logger := log.FromContext(ctx)

	var errs []error
	for _, p := range partitions {
		logger.Info("syncing queues for partition", "partition", p.Name)
		err := s.syncPartitionQueues(ctx, p)
		if err != nil {
			errs = append(errs, fmt.Errorf("syncing queues for partition %q failed: %v", p.Name, err))
		}
	}

	return errors.Join(errs...)
}

func (s *Service) syncPartitionQueues(ctx context.Context, partition *model.Partition) error {
	clientQueues, err := s.client.GetPartitionQueues(ctx, partition.Name)
	if err != nil {
		return fmt.Errorf("could not get queues for partition %s: %v", partition.Name, err)
	}

	queues := flattenQueues([]*dao.PartitionQueueDAOInfo{clientQueues})

	ids := make([]string, 0, len(queues))
	for _, q := range queues {
		ids = append(ids, q.ID)
	}

	now := time.Now().UnixNano()
	if err := s.repo.DeleteQueuesNotInIDs(ctx, partition.Name, ids, now); err != nil {
		return fmt.Errorf("could not delete queues not in IDs: %w", err)
	}

	for _, q := range queues {
		current, err := s.repo.GetQueueInPartition(ctx, partition.Name, q.ID)
		if err != nil {
			queue := &model.Queue{
				Metadata: model.Metadata{
					CreatedAtNano: now,
				},
				PartitionQueueDAOInfo: *q,
			}
			if err := s.repo.InsertQueue(ctx, queue); err != nil {
				return fmt.Errorf("could not insert queue: %w", err)
			}
			continue
		}

		current.MergeFrom(q)
		if err := s.repo.UpdateQueue(ctx, current); err != nil {
			return fmt.Errorf("could not update queue: %w", err)
		}
	}

	return nil
}

// flattenQueues returns a list of all queues in the hierarchy in a flat array.
// Usually the returned queues are a single hierarchical structure, the root queue,
// and all other queues are children queues.
func flattenQueues(qs []*dao.PartitionQueueDAOInfo) []*dao.PartitionQueueDAOInfo {
	var queues []*dao.PartitionQueueDAOInfo
	for _, q := range qs {
		queues = append(queues, q)
		if len(q.Children) > 0 {
			// update partitionName for children #148
			for i := range q.Children {
				q.Children[i].Partition = q.Partition
			}
			queues = append(queues, flattenQueues(util.ToPtrSlice(q.Children))...)
		}
	}
	return queues
}

func (s *Service) syncNodes(ctx context.Context, partitions []*model.Partition) error {
	var errs []error
	for _, p := range partitions {
		nodes, err := s.client.GetPartitionNodes(ctx, p.Name)
		if err != nil {
			return fmt.Errorf("could not get nodes for partition %s: %v", p.Name, err)
		}

		dbNodes, err := s.repo.GetLatestNodesByID(ctx, p.Name)
		if err != nil {
			return err
		}

		lookup := make(map[string]*model.Node, len(dbNodes))
		for _, n := range dbNodes {
			lookup[n.NodeID] = n
		}

		now := time.Now().UnixNano()
		for _, n := range nodes {
			current, ok := lookup[n.NodeID]
			delete(lookup, n.NodeID)
			if !ok || current.DeletedAtNano != nil { // either not exists or deleted
				node := &model.Node{
					Metadata: model.Metadata{
						ID:            ulid.Make().String(),
						CreatedAtNano: now,
					},
					Partition:   &p.Name,
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

		for _, n := range lookup {
			n.DeletedAtNano = &now
			if err := s.repo.UpdateNode(ctx, n); err != nil {
				errs = append(errs, fmt.Errorf("failed to update deleted at for node %q: %v", n.NodeID, err))
			}
		}
	}

	return errors.Join(errs...)
}

// syncApplications fetches applications for each queue and upserts them into the database
func (s *Service) syncApplications(ctx context.Context) error {
	applications, err := s.client.GetApplications(ctx, "", "")
	if err != nil {
		return fmt.Errorf("could not get applications: %v", err)
	}

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

// upsertNodeUtilizations fetches node utilizations from the Yunikorn API and inserts them into the database
func (s *Service) upsertNodeUtilizations(ctx context.Context) error {
	logger := log.FromContext(ctx)

	nus, err := s.client.GetNodeUtil(ctx)
	if err != nil {
		return fmt.Errorf("could not get node utilizations: %v", err)
	}

	err = s.workqueue.Add(func(ctx context.Context) error {
		logger.Infow("upserting node utilizations", "count", len(nus))
		return s.repo.InsertNodeUtilizations(ctx, nus)
	}, workqueue.WithJobName("upsert_node_utilizations"))
	if err != nil {
		logger.Errorf("could not add insert node utilizations job to workqueue: %v", err)
	}

	return nil
}
