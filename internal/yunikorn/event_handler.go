package yunikorn

import (
	"context"
	"encoding/json"

	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/oklog/ulid/v2"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
)

type EventHandler func(context.Context, *si.EventRecord) error

// TODO(mo-fatah):
// Experiment with the following:
// - Is there a way to see when exactly an application allocation info is going to be flushed
//   out from the scheduler? Update: Sadly no, the scheduler does not provide such information.
// - Does an application produces allocations-requests events after running? Update: Yes.

func (s *Service) handleEvent(ctx context.Context, ev *si.EventRecord) error {
	logger := log.FromContext(ctx)

	switch ev.GetType() {
	case si.EventRecord_UNKNOWN_EVENTRECORD_TYPE:
	case si.EventRecord_REQUEST:
	case si.EventRecord_APP:
		s.handleAppEvent(ctx, ev)
	case si.EventRecord_NODE:
	case si.EventRecord_QUEUE:
		s.handleQueueEvent(ctx, ev)
	case si.EventRecord_USERGROUP:
	default:
		logger.Errorf("unknown event type: %v", ev.GetType())
	}

	// TODO: Handle error
	return nil
}

// handleAppEvent handles an event of type APP.
// It will query the scheduler for the application info if the event is of the following types:
// - EventChangeType: ADD, EventChangeDetail: APP_NEW
// - EventChangeType: ADD, EventChangeDetail: DETAILS_NONE
// - EventChangeType: SET, EventChangeDetail: APP_NEW
// It will persist the application info in the database if the event is of the following types:
// - EventChangeType: SET, EventChangeDetail: APP_COMPLETED
// - EventChangeType: SET, EventChangeDetail: APP_FAILED
// - EventChangeType: REMOVE, EventChangeDetail: APP_REJECT
//
// Otherwise, it will update the application info in the cache based on the received event.
func (s *Service) handleAppEvent(ctx context.Context, ev *si.EventRecord) {
	logger := log.FromContext(ctx)

	var daoApp dao.ApplicationDAOInfo
	if err := json.Unmarshal([]byte(ev.GetState()), &daoApp); err != nil {
		logger.Errorw("Failed to unmarshal application state from event", "error", err)
		return
	}

	isNew :=
		ev.GetEventChangeType() == si.EventRecord_ADD &&
			(ev.GetEventChangeDetail() == si.EventRecord_APP_NEW || ev.GetEventChangeDetail() == si.EventRecord_DETAILS_NONE)

	var app *model.Application
	if isNew {
		app = &model.Application{
			ModelMetadata: model.ModelMetadata{
				ID:        ulid.Make().String(),
				CreatedAt: ev.TimestampNano,
			},
			ApplicationDAOInfo: daoApp,
		}

		if err := s.repo.InsertApplication(ctx, app); err != nil {
			logger.Errorf("could not insert application: %v", err)
			return
		}

		return
	}

	app, err := s.repo.GetActiveApplicationByApplicationID(ctx, daoApp.ApplicationID)
	if err != nil {
		logger.Errorf("could not get application by application id: %v", err)
		return
	}

	app.MergeFrom(daoApp)
	if ev.GetEventChangeType() == si.EventRecord_REMOVE {
		app.DeletedAt = &ev.TimestampNano
	}

	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		logger.Errorf("could not update application: %v", err)
		return
	}
}

func (s *Service) handleQueueEvent(ctx context.Context, ev *si.EventRecord) {
	logger := log.FromContext(ctx)
	logger.Debugf("adding queue event to accumulator: %v", ev)
	s.queueEventAccumulator.add(ev)
}

func (s *Service) handleQueueEvents(ctx context.Context, events []*si.EventRecord) {
	logger := log.FromContext(ctx)

	logger.Debug("Processing queue events")
	for _, event := range events {
		logger.Debugf("Event: %v", event)
	}

	s.handleQueueAddEvent(ctx)
	logger.Debug("Finished processing queue events")
}

func (s *Service) handleQueueAddEvent(ctx context.Context) {
	logger := log.FromContext(ctx)

	partitions, err := s.syncPartitions(ctx)
	if err != nil {
		logger.Errorf("could not get partitions: %v", err)
		return
	}

	if _, err := s.syncQueues(ctx, partitions); err != nil {
		logger.Errorf("could not get and sync queues: %v", err)
		return
	}
}
