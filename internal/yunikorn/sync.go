package yunikorn

import (
	"context"
	"fmt"
	"github.com/G-Research/yunikorn-history-server/internal/util"
	"sync"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/google/uuid"
)

// sync fetches the state of the applications from the Yunikorn API and upserts them into the database
func (s *Service) sync(ctx context.Context) error {
	partitions, err := s.upsertPartitions(ctx)
	if err != nil {
		return fmt.Errorf("error getting and upserting partitions: %v", err)
	}

	queues, err := s.upsertPartitionQueues(ctx, partitions)
	if err != nil {
		return fmt.Errorf("error getting and upserting queues: %v", err)
	}

	if err = s.upsertPartitionNodes(ctx, partitions); err != nil {
		return fmt.Errorf("error getting and upserting nodes: %v", err)
	}

	if err = s.upsertApplications(ctx, queues); err != nil {
		return fmt.Errorf("error getting and upserting applications: %v", err)
	}

	if err = s.upsertNodeUtilizations(ctx); err != nil {
		return fmt.Errorf("error getting and upserting node utilizations: %v", err)
	}

	if err = s.updateAppsHistory(ctx); err != nil {
		return fmt.Errorf("error updating apps history: %v", err)
	}

	return nil
}

// upsertPartitions fetches partitions from the Yunikorn API and upserts them into the database
func (s *Service) upsertPartitions(ctx context.Context) ([]*dao.PartitionInfo, error) {
	// Get partitions from Yunikorn API and upsert into DB
	partitions, err := s.client.GetPartitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get partitions: %v", err)
	}
	if err = s.repo.UpsertPartitions(ctx, partitions); err != nil {
		return nil, fmt.Errorf("could not upsert partitions: %v", err)
	}
	return partitions, nil
}

// upsertPartitionQueues fetches queues for each partition and upserts them into the database
func (s *Service) upsertPartitionQueues(ctx context.Context, partitions []*dao.PartitionInfo) ([]*dao.PartitionQueueDAOInfo, error) {
	// Create a wait group as a separate goroutine will be spawned for each partition
	wg := sync.WaitGroup{}
	wg.Add(len(partitions))

	// Protect the queues and errors slices with a mutex so multiple goroutines can safely append queues
	mutex := sync.Mutex{}

	var errs []error
	var queues []*dao.PartitionQueueDAOInfo

	processPartition := func(p *dao.PartitionInfo) {
		defer wg.Done()
		queue, err := s.client.GetPartitionQueues(ctx, p.Name)
		mutex.Lock()
		defer mutex.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("could not get queues for partition %s: %v", p.Name, err))
		} else {
			queues = append(queues, queue)
		}
	}

	for _, p := range partitions {
		go processPartition(p)
	}

	// wait for all partitions to be processed
	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get queues for some partitions: %v", errs)
	}

	queues = flattenQueues(queues)
	if err := s.repo.UpsertQueues(ctx, queues); err != nil {
		return nil, fmt.Errorf("failed to upsert queues: %v", err)
	}

	return queues, nil
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

// upsertPartitionNodes fetches nodes for each partition and upserts them into the database
func (s *Service) upsertPartitionNodes(ctx context.Context, partitions []*dao.PartitionInfo) error {
	// Create a wait group as a separate goroutine will be spawned for each partition
	wg := sync.WaitGroup{}
	wg.Add(len(partitions))

	// Protect the errors slice with a mutex so multiple goroutines can safely append queues
	mutex := sync.Mutex{}

	var errs []error

	processPartition := func(p *dao.PartitionInfo) {
		defer wg.Done()
		nodes, err := s.client.GetPartitionNodes(ctx, p.Name)
		if err != nil {
			mutex.Lock()
			errs = append(errs, fmt.Errorf("could not get nodes for partition %s: %v", p.Name, err))
			mutex.Unlock()
			return
		}
		if err = s.repo.UpsertNodes(ctx, nodes, p.Name); err != nil {
			mutex.Lock()
			errs = append(errs, fmt.Errorf("could not upsert nodes: %v", err))
			mutex.Unlock()
		}
	}

	for _, p := range partitions {
		go processPartition(p)
	}

	// wait for all partitions to be processed
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to get nodes for some partitions: %v", errs)
	}

	return nil
}

// upsertApplications fetches applications for each queue and upserts them into the database
func (s *Service) upsertApplications(ctx context.Context, queues []*dao.PartitionQueueDAOInfo) error {
	// Create a wait group as a separate goroutine will be spawned for each partition
	wg := sync.WaitGroup{}
	wg.Add(len(queues))

	// Protect the applications and errors slices with a mutex so multiple goroutines can safely append queues
	mutex := sync.Mutex{}

	var errs []error
	var apps []*dao.ApplicationDAOInfo

	processQueue := func(q *dao.PartitionQueueDAOInfo) {
		defer wg.Done()
		queueApps, err := s.client.GetApplications(ctx, q.Partition, q.QueueName)
		if err != nil {
			mutex.Lock()
			errs = append(
				errs,
				fmt.Errorf("could not get applications for partition %s, queue %s: %v", q.Partition, q.QueueName, err),
			)
			mutex.Unlock()
		} else {
			mutex.Lock()
			apps = append(apps, queueApps...)
			mutex.Unlock()
		}
	}

	// Get applications from Yunikorn API and upsert into DB
	for _, q := range queues {
		go processQueue(q)
	}

	// wait for all queues to be processed
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to get applications for some queues: %v", errs)
	}

	if err := s.repo.UpsertApplications(ctx, apps); err != nil {
		return fmt.Errorf("could not upsert applications: %v", err)
	}

	return nil
}

// upsertNodeUtilizations fetches node utilizations from the Yunikorn API and inserts them into the database
func (s *Service) upsertNodeUtilizations(ctx context.Context) error {
	nus, err := s.client.GetNodeUtil(ctx)
	if err != nil {
		return fmt.Errorf("could not get node utilizations: %v", err)
	}
	if err := s.repo.InsertNodeUtilizations(ctx, uuid.New(), nus); err != nil {
		return fmt.Errorf("could not insert node utilizations: %v", err)
	}
	return nil
}

// updateAppsHistory fetches the history of applications and containers and updates the history in the database
func (s *Service) updateAppsHistory(ctx context.Context) error {
	appsHistory, err := s.client.GetAppsHistory(ctx)
	if err != nil {
		return fmt.Errorf("could not get apps history: %v", err)
	}
	containersHistory, err := s.client.GetContainersHistory(ctx)
	if err != nil {
		return fmt.Errorf("could not get containers history: %v", err)
	}
	if err = s.repo.UpdateHistory(ctx, appsHistory, containersHistory); err != nil {
		return fmt.Errorf("could not update history: %v", err)
	}

	return nil
}
