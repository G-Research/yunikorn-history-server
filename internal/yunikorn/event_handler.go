package yunikorn

import (
	"context"

	"github.com/G-Research/yunikorn-history-server/internal/log"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/G-Research/yunikorn-scheduler-interface/lib/go/si"
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

	// We may recieve an event for a new application with the following EventChangeDetail and EventChangeType combination:
	// 1- EventChangeType: ADD, EventChangeDetail: APP_NEW
	// 2- EventChangeType: ADD, EventChangeDetail: DETAILS_NONE
	// 3- EventChangeType: SET, EventChangeDetail: APP_NEW
	// TODO(mo-fatah): Investigate if those cases are the same or different.
	// A possible difference might be combination 1 means that a brand new application was added,
	// while combination 3 might mean that the application state was set to NEW but the application
	// was already in the scheduler before with a different state.
	switch ev.GetEventChangeType() {
	case si.EventRecord_ADD:
		s.handleAppAddEvent(ctx, ev)
	case si.EventRecord_SET:
		s.handleAppSetEvent(ctx, ev)
	case si.EventRecord_REMOVE:
		s.handleAppRemoveEvent(ctx, ev)
	case si.EventRecord_NONE:
	default:
		// should be warning
		logger.Warnf("unknown event EventChangeType for an Event of type APP: %v", ev.GetEventChangeType())
	}
}

func (s *Service) handleAppAddEvent(ctx context.Context, ev *si.EventRecord) {
	logger := log.FromContext(ctx)

	switch ev.GetEventChangeDetail() {
	case si.EventRecord_APP_NEW, si.EventRecord_DETAILS_NONE:
		app, err := s.client.GetApplication(ctx, "", "", ev.GetObjectID())
		if err != nil {
			logger.Errorf("could not get application info %s from scheduler: %v\nReceived Event: %v",
				ev.GetObjectID(), err, ev)
			return
		}
		if app == nil {
			logger.Errorf("received an new application event but the application was not found in the scheduler: %s",
				ev.GetObjectID())
			return
		}
		s.appMap[ev.GetObjectID()] = app
	case si.EventRecord_APP_ALLOC:
		app, ok := s.appMap[ev.GetObjectID()]
		if !ok || app == nil {
			// should be warning
			logger.Warnf("an allocation event was received for an application without a previous ADD event: %s",
				ev.GetObjectID())
			return
		}
		resources := make(map[string]int64)
		for res, quantity := range ev.GetResource().GetResources() {
			resources[res] = quantity.GetValue()
		}
		alloc := &dao.AllocationDAOInfo{
			AllocationKey:    ev.GetReferenceID(),
			ResourcePerAlloc: resources,
			// AllocationTime might be different from the Event.TimestampNano
			// For now, we will use the Event.TimestampNano as the AllocationTime but this is not factually correct
			AllocationTime: ev.GetTimestampNano(),
		}
		app.Allocations = append(app.Allocations, alloc)
	case si.EventRecord_APP_REQUEST:
		app, ok := s.appMap[ev.GetObjectID()]
		if !ok || app == nil {
			// should be warning
			logger.Warnf(
				"an allocation request event was received for an application without a previous ADD event: %s",
				ev.GetObjectID())
			return
		}
		requestedResources := make(map[string]int64)
		for res, quantity := range ev.GetResource().GetResources() {
			requestedResources[res] = quantity.GetValue()
		}
		ask := &dao.AllocationAskDAOInfo{
			// As per documentation, the ReferenceID in an APP_REQUEST event is the RequestID
			// but the AllocationAskDAOInfo doesn't have a RequestID field. What does the ReferenceID represent here?
			// Assigned for now to the AllocationKey field.
			AllocationKey:    ev.GetReferenceID(),
			ResourcePerAlloc: requestedResources,
			// RequestTime might be different from the Event.TimestampNano
			// For now, we will use the Event.TimestampNano as the RequestTime but this is not factually correct
			RequestTime: ev.GetTimestampNano(),
		}
		app.Requests = append(app.Requests, ask)
	default:
		// should be warning
		logger.Warnf("unknown event EventChangeDetail type for an Event of type APP: %v",
			ev.GetEventChangeDetail())
	}
}

// Usually, we receive a SET event for an application when the application state changes.
// Except for the REJECT state, which is received as a REMOVE event.
func (s *Service) handleAppSetEvent(ctx context.Context, ev *si.EventRecord) {
	logger := log.FromContext(ctx)

	switch ev.GetEventChangeDetail() {
	case si.EventRecord_APP_NEW:
		app, err := s.client.GetApplication(ctx, "", "", ev.GetObjectID())
		if err != nil {
			logger.Errorf(
				"could not get application info %s from scheduler: %v\nReceived Event: %v",
				ev.GetObjectID(), err, ev)
			return
		}
		if app == nil {
			logger.Errorf(
				"received an new application event but the application was not found in the scheduler: %s",
				ev.GetObjectID())
			return
		}
		s.appMap[ev.GetObjectID()] = app
	case si.EventRecord_APP_ACCEPTED,
		si.EventRecord_APP_RUNNING,
		si.EventRecord_APP_COMPLETING, si.EventRecord_APP_COMPLETED,
		si.EventRecord_APP_FAILING, si.EventRecord_APP_FAILED,
		si.EventRecord_APP_RESUMING, si.EventRecord_APP_EXPIRED:
		state := si.EventRecord_ChangeDetail_name[int32(ev.GetEventChangeDetail())]
		app, ok := s.appMap[ev.GetObjectID()]
		if !ok || app == nil {
			// should be warning
			logger.Warnf("an application state change of type %s was "+
				"received for an application without a previous ADD event: %s",
				state, ev.GetObjectID())
			return
		}
		app.StateLog = append(app.StateLog, &dao.StateDAOInfo{
			Time:             ev.GetTimestampNano(),
			ApplicationState: state,
		})
		app.State = state
		// Insert the application into the DB once it is completed or failed
		// should we include the EXPIRED state?
		if ev.GetEventChangeDetail() == si.EventRecord_APP_COMPLETED ||
			ev.GetEventChangeDetail() == si.EventRecord_APP_FAILED {

			if err := s.repo.UpsertApplications(ctx, []*dao.ApplicationDAOInfo{app}); err != nil {
				logger.Errorf("could not insert application into DB: %v", err)
			}
		}
	default:
		// should be warning
		logger.Warnf("unknown event EventChangeDetail type for an Event of type APP: %v",
			ev.GetEventChangeDetail())
	}
}

func (s *Service) handleAppRemoveEvent(ctx context.Context, ev *si.EventRecord) {
	logger := log.FromContext(ctx)

	switch ev.GetEventChangeDetail() {
	case si.EventRecord_DETAILS_NONE:
		// Should we reinsert the application into the DB in case we didn't a terminal state change event (e.g. completed)?
		delete(s.appMap, ev.GetObjectID())
	case si.EventRecord_APP_REJECT:
		app, ok := s.appMap[ev.GetObjectID()]
		if !ok || app == nil {
			// should be warning
			logger.Warnf(
				"an application rejection event was received for an "+
					"application without a previous ADD event: %s",
				ev.GetObjectID())
			return
		}
		state := si.EventRecord_ChangeDetail_name[int32(ev.GetEventChangeDetail())]
		app.StateLog = append(app.StateLog, &dao.StateDAOInfo{
			Time:             ev.GetTimestampNano(),
			ApplicationState: state,
		})
		app.State = state
		if err := s.repo.UpsertApplications(ctx, []*dao.ApplicationDAOInfo{app}); err != nil {
			logger.Errorf("could not insert application into DB: %v", err)
		}
		// should we delete the application from the cache or it is guaranteed to recieve a REMOVE with DETAILS_NONE event?
	case si.EventRecord_ALLOC_CANCEL, si.EventRecord_ALLOC_TIMEOUT,
		si.EventRecord_ALLOC_REPLACED, si.EventRecord_ALLOC_PREEMPT,
		si.EventRecord_ALLOC_NODEREMOVED, si.EventRecord_APP_REQUEST,
		si.EventRecord_REQUEST_TIMEOUT, si.EventRecord_REQUEST_CANCEL:
		// Ignored for now
	default:
		// should be warning
		logger.Warnf("unknown event EventChangeDetail type for an Event of type APP: %v",
			ev.GetEventChangeDetail())
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
