package yunikorn

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/run"

	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"github.com/G-Research/yunikorn-history-server/internal/workqueue"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"

	"github.com/G-Research/yunikorn-history-server/internal/log"
)

type Service struct {
	repo            repository.Repository
	eventRepository repository.EventRepository
	client          Client
	// eventHandler is a function that handles events from the Yunikorn event stream.
	eventHandler EventHandler
	// appMap is a map of application IDs to their respective DAOs.
	appMap map[string]*dao.ApplicationDAOInfo
	// syncInterval is the interval at which the service will sync the state of the applications with the Yunikorn API.
	syncInterval time.Duration
	// workqueue processes jobs which store data in database during data sync and retries them with exponential backoff.
	workqueue *workqueue.WorkQueue
}

type Option func(*Service)

func WithSyncInterval(interval time.Duration) Option {
	return func(s *Service) {
		s.syncInterval = interval
	}
}

func NewService(repository repository.Repository, eventRepository repository.EventRepository, client Client, opts ...Option) *Service {
	s := &Service{
		repo:            repository,
		eventRepository: eventRepository,
		client:          client,
		appMap:          make(map[string]*dao.ApplicationDAOInfo),
		syncInterval:    5 * time.Minute,
		workqueue:       workqueue.NewWorkQueue(workqueue.WithName("yunikorn_data_sync")),
	}
	s.eventHandler = s.handleEvent
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
		return s.runEventCollector(ctx)
	}, func(err error) {},
	)

	g.Add(func() error {
		return s.runDataSync(ctx)
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

// RunDataSync starts the data sync process which periodically syncs the state of the applications with the Yunikorn API.
func (s *Service) runDataSync(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger = logger.With("component", "yunikorn_data_sync")
	ctx = log.ToContext(ctx, logger)

	logger.Info("starting yunikorn data sync")

	if err := s.sync(ctx); err != nil {
		logger.Errorf("error syncing data with yunikorn api: %v", err)
	}

	// if sync interval is 0, sync data once and return
	if s.syncInterval == 0 {
		logger.Info("sync interval is not configured, shutting down data sync")
		return nil
	}

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Warn("shutting down data sync")
			return nil
		case <-ticker.C:
			logger.Info("syncing data with yunikorn api")
			if err := s.sync(ctx); err != nil {
				logger.Errorf("error syncing data with yunikorn api: %v", err)
			}
		}
	}
}
