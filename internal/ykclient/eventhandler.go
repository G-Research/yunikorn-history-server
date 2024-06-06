package ykclient

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
)

// TODO(mo-fatah):
// Experiment with the following:
// - Is there a way to see when exactly an application allocation info is going to be flushed
//   out from the scheduler? Update: Sadly no, the scheduler does not provide such information.
// - Does an application produces allocations-requests events after running? Update: Yes.

func (c *Client) handleEvent(ctx context.Context, ev *si.EventRecord) {
	switch ev.GetType() {
	case si.EventRecord_UNKNOWN_EVENTRECORD_TYPE:
	case si.EventRecord_REQUEST:
	case si.EventRecord_APP:
		c.handleAppEvent(ctx, ev)
	case si.EventRecord_NODE:
	case si.EventRecord_QUEUE:
	case si.EventRecord_USERGROUP:
	default:
		fmt.Fprintf(os.Stderr, "Unknown event type: %v\n", ev.GetType())
	}
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
func (c *Client) handleAppEvent(ctx context.Context, ev *si.EventRecord) {
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
		c.handleAppAddEvent(ctx, ev)
	case si.EventRecord_SET:
		c.handleAppSetEvent(ctx, ev)
	case si.EventRecord_REMOVE:
		c.handleAppRemoveEvent(ctx, ev)
	case si.EventRecord_NONE:
	default:
		// should be warning
		fmt.Fprintf(os.Stderr, "Unknown event EventChangeType for an Event of type APP: %v\n", ev.GetEventChangeType())
	}
}

func (c *Client) handleAppAddEvent(ctx context.Context, ev *si.EventRecord) {
	switch ev.GetEventChangeDetail() {
	case si.EventRecord_APP_NEW, si.EventRecord_DETAILS_NONE:
		app, err := c.GetApplication(ctx, "", "", ev.GetObjectID())
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get application info %s from scheduler: %v\nReceived Event: %v\n",
				ev.GetObjectID(), err, ev)
			return
		}
		if app == nil {
			fmt.Fprintf(os.Stderr, "received an new application event but the application was not found in the scheduler: %s\n",
				ev.GetObjectID())
			return
		}
		c.appMap[ev.GetObjectID()] = app
	case si.EventRecord_APP_ALLOC:
		app, ok := c.appMap[ev.GetObjectID()]
		if !ok || app == nil {
			// should be warning
			fmt.Fprintf(os.Stderr, "an allocation event was received for an application without a previous ADD event: %s\n",
				ev.GetObjectID())
			return
		}
		resources := make(map[string]int64)
		for res, quantity := range ev.GetResource().GetResources() {
			resources[res] = quantity.GetValue()
		}
		alloc := &dao.AllocationDAOInfo{
			AllocationID:     ev.GetReferenceID(),
			ResourcePerAlloc: resources,
			// AllocationTime might be different from the Event.TimestampNano
			// For now, we will use the Event.TimestampNano as the AllocationTime but this is not factually correct
			AllocationTime: ev.GetTimestampNano(),
		}
		app.Allocations = append(app.Allocations, alloc)
	case si.EventRecord_APP_REQUEST:
		app, ok := c.appMap[ev.GetObjectID()]
		if !ok || app == nil {
			// should be warning
			fmt.Fprintf(os.Stderr,
				"an allocation request event was received for an application without a previous ADD event: %s\n",
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
		fmt.Fprintf(os.Stderr, "Unknown event EventChangeDetail type for an Event of type APP: %v\n",
			ev.GetEventChangeDetail())
	}
}

// Usually, we receive a SET event for an application when the application state changes.
// Except for the REJECT state, which is received as a REMOVE event.
func (c *Client) handleAppSetEvent(ctx context.Context, ev *si.EventRecord) {
	switch ev.GetEventChangeDetail() {
	case si.EventRecord_APP_NEW:
		app, err := c.GetApplication(ctx, "", "", ev.GetObjectID())
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get application info %s from scheduler: %v\nReceived Event: %v\n",
				ev.GetObjectID(), err, ev)
			return
		}
		if app == nil {
			fmt.Fprintf(os.Stderr, "received an new application event but the application was not found in the scheduler: %s\n",
				ev.GetObjectID())
			return
		}
		c.appMap[ev.GetObjectID()] = app
	case si.EventRecord_APP_ACCEPTED,
		si.EventRecord_APP_STARTING, si.EventRecord_APP_RUNNING,
		si.EventRecord_APP_COMPLETING, si.EventRecord_APP_COMPLETED,
		si.EventRecord_APP_FAILING, si.EventRecord_APP_FAILED,
		si.EventRecord_APP_RESUMING, si.EventRecord_APP_EXPIRED:
		state := si.EventRecord_ChangeDetail_name[int32(ev.GetEventChangeDetail())]
		app, ok := c.appMap[ev.GetObjectID()]
		if !ok || app == nil {
			// should be warning
			fmt.Fprintf(os.Stderr,
				"an application state change of type %s was received for an application without a previous ADD event: %s\n",
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

			if err := c.repo.UpsertApplications(ctx, []*dao.ApplicationDAOInfo{app}); err != nil {
				fmt.Fprintf(os.Stderr, "could not insert application into DB: %v\n", err)
			}
		}
	default:
		// should be warning
		fmt.Fprintf(os.Stderr, "Unknown event EventChangeDetail type for an Event of type APP: %v\n",
			ev.GetEventChangeDetail())
	}
}

func (c *Client) handleAppRemoveEvent(ctx context.Context, ev *si.EventRecord) {
	switch ev.GetEventChangeDetail() {
	case si.EventRecord_DETAILS_NONE:
		// Should we reinsert the application into the DB in case we didn't a terminal state change event (e.g. completed)?
		delete(c.appMap, ev.GetObjectID())
	case si.EventRecord_APP_REJECT:
		app, ok := c.appMap[ev.GetObjectID()]
		if !ok || app == nil {
			// should be warning
			fmt.Fprintf(os.Stderr,
				"an application rejection event was received for an application without a previous ADD event: %s\n",
				ev.GetObjectID())
			return
		}
		state := si.EventRecord_ChangeDetail_name[int32(ev.GetEventChangeDetail())]
		app.StateLog = append(app.StateLog, &dao.StateDAOInfo{
			Time:             ev.GetTimestampNano(),
			ApplicationState: state,
		})
		app.State = state
		if err := c.repo.UpsertApplications(ctx, []*dao.ApplicationDAOInfo{app}); err != nil {
			fmt.Fprintf(os.Stderr, "could not insert application into DB: %v\n", err)
		}
		// should we delete the application from the cache or it is guaranteed to recieve a REMOVE with DETAILS_NONE event?
	case si.EventRecord_ALLOC_CANCEL, si.EventRecord_ALLOC_TIMEOUT,
		si.EventRecord_ALLOC_REPLACED, si.EventRecord_ALLOC_PREEMPT,
		si.EventRecord_ALLOC_NODEREMOVED, si.EventRecord_APP_REQUEST,
		si.EventRecord_REQUEST_TIMEOUT, si.EventRecord_REQUEST_CANCEL:
		// Ignored for now
	default:
		// should be warning
		fmt.Fprintf(os.Stderr, "Unknown event EventChangeDetail type for an Event of type APP: %v\n",
			ev.GetEventChangeDetail())
	}
}
