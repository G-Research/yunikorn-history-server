package yunikorn

import (
	"context"
	"fmt"
	"sync"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"

	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"
	"github.com/G-Research/yunikorn-history-server/internal/workqueue"
)

// sync fetches the state of the applications from the Yunikorn API and upserts them into the database
func (s *Service) sync(ctx context.Context) error {
	partitions, err := s.syncPartitions(ctx)
	if err != nil {
		return fmt.Errorf("error getting and upserting partitions: %v", err)
	}

	var mu sync.Mutex
	var allErrs []error
	addErr := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		allErrs = append(allErrs, err)
	}

	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		defer wg.Done()
		queues, err := s.syncQueues(ctx, partitions)
		if err != nil {
			addErr(fmt.Errorf("error getting and upserting queues: %v", err))
			return
		}

		if err = s.upsertApplications(ctx, queues); err != nil {
			addErr(fmt.Errorf("error getting and upserting applications: %v", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err = s.upsertPartitionNodes(ctx, partitions); err != nil {
			addErr(fmt.Errorf("error getting and upserting nodes: %v", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err = s.upsertNodeUtilizations(ctx); err != nil {
			addErr(fmt.Errorf("error getting and upserting node utilizations: %v", err))
		}
	}()

	go func() {
		defer wg.Done()
		if err = s.updateAppsHistory(ctx); err != nil {
			addErr(fmt.Errorf("error updating apps history: %v", err))
		}
	}()

	wg.Wait()

	if len(allErrs) > 0 {
		return fmt.Errorf("some errors encountered while syncing data: %v", allErrs)
	}

	return nil
}

// syncPartitions fetches partitions from the Yunikorn API and syncs them into the database
func (s *Service) syncPartitions(ctx context.Context) ([]*dao.PartitionInfo, error) {
	logger := log.FromContext(ctx)
	// Get partitions from Yunikorn API and upsert into DB
	partitions, err := s.client.GetPartitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get partitions: %v", err)
	}

	err = s.workqueue.Add(func(ctx context.Context) error {
		logger.Infow("syncing partitions", "count", len(partitions))
		err := s.repo.UpsertPartitions(ctx, partitions)
		if err != nil {
			return fmt.Errorf("could not upsert partitions: %w", err)
		}
		// Delete partitions that are not present in the API response
		return s.repo.DeletePartitions(ctx, partitions)
	}, workqueue.WithJobName("sync_partitions"))
	if err != nil {
		logger.Errorf("could not add sync_partitions job to workqueue: %v", err)
	}

	return partitions, nil
}

// syncQueues fetches queues for each partition and upserts them into the database
func (s *Service) syncQueues(ctx context.Context, partitions []*dao.PartitionInfo) ([]*dao.PartitionQueueDAOInfo, error) {
	logger := log.FromContext(ctx)

	errs := make(chan error, len(partitions))
	partitionQueues := make(chan []*dao.PartitionQueueDAOInfo, len(partitions))
	var wg sync.WaitGroup
	wg.Add(len(partitions))

	for _, partition := range partitions {
		go func() {
			defer wg.Done()
			logger.Infow("syncing queues for partition", "partition", partition.Name)
			queues, err := s.syncPartitionQueues(ctx, partition)
			if err != nil {
				errs <- err
				return
			}
			partitionQueues <- queues
		}()
	}
	wg.Wait()
	close(errs)
	close(partitionQueues)

	var syncErrors []error
	for err := range errs {
		syncErrors = append(syncErrors, err)
	}

	if len(syncErrors) > 0 {
		return nil, fmt.Errorf("some errors encountered while syncing queues: %v", syncErrors)
	}

	// retrieve all queues from the channel
	var queues []*dao.PartitionQueueDAOInfo
	for qs := range partitionQueues {
		queues = append(queues, qs...)
	}
	return queues, nil
}

func (s *Service) syncPartitionQueues(ctx context.Context, partition *dao.PartitionInfo) ([]*dao.PartitionQueueDAOInfo, error) {
	// Fetch partition queues from the YuniKorn API
	queue, err := s.client.GetPartitionQueues(ctx, partition.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve queues from YuniKorn API: %w", err)
	}

	// Attempt to update the queue; if it fails, try adding it instead
	if err := s.repo.UpdateQueue(ctx, queue); err != nil {
		if addErr := s.repo.AddQueues(ctx, nil, []*dao.PartitionQueueDAOInfo{queue}); addErr != nil {
			return nil, fmt.Errorf("failed to add new queue: %w", addErr)
		}
	}

	queues := flattenQueues([]*dao.PartitionQueueDAOInfo{queue})
	// Find candidates for deletion
	deleteCandidates, err := s.findQueueDeleteCandidates(ctx, partition, queues)
	if err != nil {
		return nil, fmt.Errorf("failed to find delete candidates: %w", err)
	}

	// Delete the identified queues
	if err := s.repo.DeleteQueues(ctx, deleteCandidates); err != nil {
		return nil, fmt.Errorf("failed to delete queues: %w", err)
	}

	return queues, nil
}

func (s *Service) findQueueDeleteCandidates(
	ctx context.Context,
	partition *dao.PartitionInfo,
	apiQueues []*dao.PartitionQueueDAOInfo) ([]*model.PartitionQueueDAOInfo, error) {

	// Fetch queues from the database for the given partition
	queuesInDB, err := s.repo.GetQueuesPerPartition(ctx, partition.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve queues from DB: %w", err)
	}

	apiQueueMap := make(map[string]*dao.PartitionQueueDAOInfo)
	for _, queue := range apiQueues {
		apiQueueMap[queue.QueueName] = queue
	}

	// Identify queues in the database that are not present in the API response
	var deleteCandidates []*model.PartitionQueueDAOInfo
	for _, dbQueue := range queuesInDB {
		if _, found := apiQueueMap[dbQueue.QueueName]; !found {
			deleteCandidates = append(deleteCandidates, dbQueue)
		}
	}

	return deleteCandidates, nil
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

// upsertPartitionNodes fetches nodes for each partition and upserts them into the database
func (s *Service) upsertPartitionNodes(ctx context.Context, partitions []*dao.PartitionInfo) error {
	logger := log.FromContext(ctx)

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
		err = s.workqueue.Add(func(ctx context.Context) error {
			logger.Infow("upserting nodes for partition", "count", len(nodes), "partition", p.Name)
			return s.repo.UpsertNodes(ctx, nodes, p.Name)
		}, workqueue.WithJobName(fmt.Sprintf("upsert_nodes_for_partition_%s", p.Name)))
		if err != nil {
			logger.Errorf("could not add upsert nodes for partition %s job to workqueue: %v", p.Name, err)
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
	logger := log.FromContext(ctx)

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

	err := s.workqueue.Add(func(ctx context.Context) error {
		logger.Infow("upserting applications", "count", len(apps))
		return s.repo.UpsertApplications(ctx, apps)
	}, workqueue.WithJobName("upsert_applications"))
	if err != nil {
		logger.Errorf("could not add upsert applications job to workqueue: %v", err)
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

// updateAppsHistory fetches the history of applications and containers and updates the history in the database
func (s *Service) updateAppsHistory(ctx context.Context) error {
	logger := log.FromContext(ctx)

	appsHistory, err := s.client.GetAppsHistory(ctx)
	if err != nil {
		return fmt.Errorf("could not get apps history: %v", err)
	}
	containersHistory, err := s.client.GetContainersHistory(ctx)
	if err != nil {
		return fmt.Errorf("could not get containers history: %v", err)
	}

	err = s.workqueue.Add(func(ctx context.Context) error {
		logger.Infow("updating apps history", "count", len(appsHistory))
		return s.repo.UpdateHistory(ctx, appsHistory, containersHistory)
	}, workqueue.WithJobName("update_apps_history"))
	if err != nil {
		logger.Errorf("could not add update apps history job to workqueue: %v", err)
	}

	return nil
}
