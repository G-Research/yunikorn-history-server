package yunikorn

import (
	"context"
	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"time"

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
	// eventCollectorRunning is a flag that indicates whether the event collector is running.
	eventCollectorRunning bool
}

func NewService(repository repository.Repository, eventRepository repository.EventRepository, client Client) *Service {
	s := &Service{
		repo:            repository,
		eventRepository: eventRepository,
		client:          client,
		appMap:          make(map[string]*dao.ApplicationDAOInfo),
		syncInterval:    5 * time.Minute,
	}
	s.eventHandler = s.handleEvent
	return s
}

func (s *Service) Shutdown() {
	s.eventCollectorRunning = false
}

// RunEventCollector starts the event stream client which processes events from the Yunikorn event stream.
// It maintains a persistent connection to the Yunikorn event stream endpoint, and retries in case of any errors.
func (s *Service) RunEventCollector(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger = logger.With("component", "yunikorn_event_collector")
	ctx = log.ToContext(ctx, logger)

	s.eventCollectorRunning = true

	logger.Info("starting yunikorn event stream client")
	for {
		if !s.eventCollectorRunning {
			logger.Warn("shutting down yunikorn event stream client")
			return nil
		}
		err := s.ProcessEvents(ctx)
		if err != nil {
			logger.Errorf("error processing yunikorn events: %v", err)
		}
		logger.Info("reconnecting yunikorn event stream client")
		time.Sleep(2 * time.Second)
	}
}

// RunDataSync starts the data sync process which periodically syncs the state of the applications with the Yunikorn API.
func (s *Service) RunDataSync(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger = logger.With("component", "yunikorn_data_sync")
	ctx = log.ToContext(ctx, logger)

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	logger.Info("starting yunikorn data sync")

	if err := s.sync(ctx); err != nil {
		logger.Errorf("error syncing data with yunikorn api: %v", err)
	}

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
