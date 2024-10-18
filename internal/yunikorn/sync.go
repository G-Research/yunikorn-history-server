package yunikorn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/oklog/ulid/v2"

	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/internal/workqueue"
)

// syncPartitions fetches partitions from the Yunikorn API and syncs them into the database
func (s *Service) syncPartitions(ctx context.Context) ([]*model.Partition, error) {
	logger := log.FromContext(ctx)
	// Get partitions from Yunikorn API and upsert into DB
	partitions, err := s.client.GetPartitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get partitions: %v", err)
	}

	current, err := s.repo.GetLatestPartitionsGroupedByName(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get latest partitions: %v", err)
	}

	lookup := make(map[string]*model.Partition, len(current))
	for _, p := range current {
		lookup[p.Name] = p
	}

	now := time.Now().UnixNano()
	allPartitions := make([]*model.Partition, 0, len(partitions))
	for _, p := range partitions {
		current, ok := lookup[p.Name]
		delete(lookup, p.Name)
		if !ok || current.DeletedAtNano != nil { // either not exists or deleted
			partition := &model.Partition{
				Metadata: model.Metadata{
					ID:            ulid.Make().String(),
					CreatedAtNano: now,
				},
				PartitionInfo: *p,
			}
			allPartitions = append(allPartitions, partition)
			if err := s.repo.InsertPartition(ctx, partition); err != nil {
				logger.Errorf("could not create partition %s: %v", p.Name, err)
			}
			continue
		}

		current.MergeFrom(p)
		allPartitions = append(allPartitions, current)
		if err := s.repo.UpdatePartition(ctx, current); err != nil {
			logger.Errorf("could not update partition %s: %v", p.Name, err)
		}
	}

	for _, p := range lookup {
		p.DeletedAtNano = &now
		if err := s.repo.UpdatePartition(ctx, p); err != nil {
			logger.Errorf("failed to update deleted at for partition %q: %v", p.Name, err)
		}
	}

	return allPartitions, nil
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

	dbQueues, err := s.repo.GetQueuesInPartition(ctx, partition.Name)
	if err != nil {
		return err
	}

	lookup := make(map[string]*model.Queue, len(dbQueues))
	for _, q := range dbQueues {
		lookup[q.QueueName] = q
	}

	now := time.Now().UnixNano()
	var errs []error
	for _, q := range queues {
		current, ok := lookup[q.QueueName]
		delete(lookup, q.QueueName)
		if !ok || current.DeletedAtNano != nil { // either not exists or deleted
			queue := &model.Queue{
				Metadata: model.Metadata{
					ID:            ulid.Make().String(),
					CreatedAtNano: now,
				},
				PartitionQueueDAOInfo: *q,
			}
			if err := s.repo.InsertQueue(ctx, queue); err != nil {
				errs = append(errs, fmt.Errorf("could not insert queue %s: %v", q.QueueName, err))
			}
			continue
		}

		current.MergeFrom(q)
		if err := s.repo.UpdateQueue(ctx, current); err != nil {
			errs = append(errs, err)
		}
	}

	for _, q := range lookup {
		q.DeletedAtNano = &now
		if err := s.repo.UpdateQueue(ctx, q); err != nil {
			errs = append(errs, fmt.Errorf("failed to update deleted at for queue %q: %v", q.QueueName, err))
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

	dbApplications, err := s.repo.GetLatestApplicationsByApplicationID(ctx)
	if err != nil {
		return fmt.Errorf("could not get latest applications: %v", err)
	}

	lookup := make(map[string]*model.Application, len(dbApplications))
	for _, a := range dbApplications {
		lookup[a.ApplicationID] = a
	}

	now := time.Now().UnixNano()
	var errs []error
	for _, a := range applications {
		current, ok := lookup[a.ApplicationID]
		delete(lookup, a.ApplicationID)
		if !ok || current.DeletedAtNano != nil { // either not exists or deleted
			application := &model.Application{
				Metadata: model.Metadata{
					ID:            ulid.Make().String(),
					CreatedAtNano: now,
				},
				ApplicationDAOInfo: *a,
			}
			if err := s.repo.InsertApplication(ctx, application); err != nil {
				errs = append(errs, fmt.Errorf("could not insert application %s: %v", a.ApplicationID, err))
			}
			continue
		}

		current.MergeFrom(a)
		if err := s.repo.UpdateApplication(ctx, current); err != nil {
			errs = append(errs, fmt.Errorf("could not update application %s: %v", a.ApplicationID, err))
		}
	}

	for _, a := range lookup {
		a.DeletedAtNano = &now
		if err := s.repo.UpdateApplication(ctx, a); err != nil {
			errs = append(errs, fmt.Errorf("failed to update deleted at for application %q: %v", a.ApplicationID, err))
		}
	}

	return errors.Join(errs...)
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
