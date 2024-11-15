package yunikorn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/oklog/run"

	"github.com/G-Research/unicorn-history-server/internal/database/repository"
	"github.com/G-Research/unicorn-history-server/internal/log"
)

type Service struct {
	repo            repository.Repository
	eventRepository repository.EventRepository
	client          Client
	// eventHandler is a function that handles events from the Yunikorn event stream.
	eventHandler EventHandler
	// appMap is a map of application IDs to their respective DAOs.
	appMap map[string]*dao.ApplicationDAOInfo
}

type Option func(*Service)

func NewService(repository repository.Repository, eventRepository repository.EventRepository, client Client, opts ...Option) *Service {
	s := &Service{
		repo:            repository,
		eventRepository: eventRepository,
		client:          client,
		appMap:          make(map[string]*dao.ApplicationDAOInfo),
	}
	s.eventHandler = s.handleEvent
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) Run(ctx context.Context) error {
	g := run.Group{}

	fullState, err := s.client.GetFullStateDump(ctx)
	if err != nil {
		return fmt.Errorf("could not get full state dump: %v", err)
	}

	if err := s.syncPartitions(ctx, fullState.Partitions); err != nil {
		return fmt.Errorf("error syncing partitions: %v", err)
	}
	if err := s.syncQueues(ctx, fullState.Queues); err != nil {
		return fmt.Errorf("error syncing queues: %v", err)
	}
	if err := s.syncApplications(ctx, fullState.Applications); err != nil {
		return fmt.Errorf("error syncing applications: %v", err)
	}
	if err := s.syncNodes(ctx, fullState.Nodes); err != nil {
		return fmt.Errorf("error syncing nodes: %v", err)
	}
	if err := s.syncAppHistory(ctx, fullState.AppHistory); err != nil {
		return fmt.Errorf("error syncing app history: %v", err)
	}
	if err := s.syncContainerHistory(ctx, fullState.ContainerHistory); err != nil {
		return fmt.Errorf("error syncing container history: %v", err)
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
