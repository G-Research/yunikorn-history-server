package webservice

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"github.com/G-Research/yunikorn-history-server/internal/health"
	"github.com/G-Research/yunikorn-history-server/internal/log"
)

type WebService struct {
	server          *http.Server
	repository      repository.Repository
	eventRepository repository.EventRepository
	healthService   health.Interface
	config          config.YHSConfig
}

func NewWebService(
	cfg config.YHSConfig,
	repository repository.Repository,
	eventRepository repository.EventRepository,
	healthService health.Interface,
) *WebService {
	return &WebService{
		server: &http.Server{
			Addr:        fmt.Sprintf(":%d", cfg.Port),
			ReadTimeout: 30 * time.Second,
		},
		repository:      repository,
		eventRepository: eventRepository,
		healthService:   healthService,
		config:          cfg,
	}
}

// Start performs a blocking call to start the REST API server.
func (ws *WebService) Start(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger = logger.With("component", "webservice")
	ctx = log.ToContext(ctx, logger)

	ws.init(ctx)

	logger.Infof("starting webservice on %s", ws.server.Addr)
	return ws.server.ListenAndServe()
}

func (ws *WebService) Shutdown(ctx context.Context) error {
	logger := log.FromContext(ctx)

	logger.Warnw("shutting down webservice", "component", "webservice")
	return ws.server.Shutdown(ctx)
}
