package yunikorn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/G-Research/yunikorn-scheduler-interface/lib/go/si"
	"github.com/oklog/run"

	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/internal/workqueue"
)

type Service struct {
	repo            repository.Repository
	eventRepository repository.EventRepository
	client          Client
	// eventHandler is a function that handles events from the Yunikorn event stream.
	eventHandler EventHandler
	// partitionAccumulator accumulates new queue events and synchronizes partitions after a certain interval.
	partitionAccumulator *accumulator
	// appMap is a map of application IDs to their respective DAOs.
	appMap map[string]*dao.ApplicationDAOInfo
	// workqueue processes jobs which store data in database during data sync and retries them with exponential backoff.
	workqueue *workqueue.WorkQueue
}

type Option func(*Service)

func NewService(repository repository.Repository, eventRepository repository.EventRepository, client Client, opts ...Option) *Service {
	s := &Service{
		repo:            repository,
		eventRepository: eventRepository,
		client:          client,
		appMap:          make(map[string]*dao.ApplicationDAOInfo),
		workqueue:       workqueue.NewWorkQueue(workqueue.WithName("yunikorn_data_sync")),
	}
	s.eventHandler = s.handleEvent
	s.partitionAccumulator = newAccumulator(
		func(ctx context.Context, event []*si.EventRecord) {
			logger := log.FromContext(ctx)
			_, err := s.syncPartitions(ctx)
			if err != nil {
				logger.Errorf("error syncing partitions: %v", err)
				return
			}
		},
		2*time.Second,
	)
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) Run(ctx context.Context) error {
	g := run.Group{}

	g.Add(func() error {
		return s.workqueue.Run(ctx)
	}, func(err error) {},
	)

	g.Add(func() error {
		return s.partitionAccumulator.run(ctx)
	}, func(err error) {},
	)

	partitions, err := s.syncPartitions(ctx)
	if err != nil {
		return err
	}
	if err := s.syncQueues(ctx, partitions); err != nil {
		return fmt.Errorf("error syncing queues: %v", err)
	}
	if err := s.syncApplications(ctx); err != nil {
		return fmt.Errorf("error syncing applications: %v", err)
	}
	if err := s.upsertPartitionNodes(ctx, partitions); err != nil {
		return fmt.Errorf("error upserting partition nodes: %v", err)
	}
	if err := s.upsertNodeUtilizations(ctx); err != nil {
		return fmt.Errorf("error upserting node utilizations: %v", err)
	}
	if err := s.syncHistory(ctx); err != nil {
		return fmt.Errorf("error syncing app and container history: %v", err)
	}

	g.Add(func() error {
		return s.runEventCollector(ctx)
	}, func(err error) {},
	)

	return g.Run()
}

// RunEventCollector starts the event stream client which processes events from the Yunikorn event stream.
// It maintains a persistent connection to the Yunikorn event stream endpoint, and retries in case of any errors.
func (s *Service) runEventCollector(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger = logger.With("component", "yunikorn_event_collector")
	ctx = log.ToContext(ctx, logger)

	logger.Info("starting yunikorn event stream client")
	for {
		err := s.ProcessEvents(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Warn("shutting down yunikorn event stream client")
				return nil
			}
			logger.Errorf("error processing yunikorn events: %v", err)
		}
		logger.Info("reconnecting yunikorn event stream client")
		time.Sleep(2 * time.Second)
	}
}
