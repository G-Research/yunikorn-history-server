package yunikorn

import (
	"context"
	"encoding/json"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/G-Research/yunikorn-scheduler-interface/lib/go/si"

	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/G-Research/yunikorn-history-server/internal/model"
)

type EventHandler func(context.Context, *si.EventRecord) error

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

	return nil
}

func (s *Service) handleAppEvent(ctx context.Context, ev *si.EventRecord) {
	logger := log.FromContext(ctx)

	var daoApp dao.ApplicationDAOInfo
	if err := json.Unmarshal([]byte(ev.GetState()), &daoApp); err != nil {
		logger.Errorw("failed to unmarshal application state from event", "error", err)
		return
	}

	isNew := ev.GetEventChangeType() == si.EventRecord_ADD &&
		(ev.GetEventChangeDetail() == si.EventRecord_APP_NEW || ev.GetEventChangeDetail() == si.EventRecord_DETAILS_NONE)

	var app *model.Application
	if isNew {
		app = &model.Application{
			Metadata: model.Metadata{
				CreatedAtNano: ev.TimestampNano,
			},
			ApplicationDAOInfo: daoApp,
		}

		if err := s.repo.InsertApplication(ctx, app); err != nil {
			logger.Errorf("could not insert application: %v", err)
			return
		}

		return
	}

	app, err := s.repo.GetApplicationByID(ctx, daoApp.ID)
	if err != nil {
		logger.Errorf("could not get application by application id: %v", err)
		return
	}

	app.MergeFrom(&daoApp)
	if ev.GetEventChangeType() == si.EventRecord_REMOVE {
		app.DeletedAtNano = &ev.TimestampNano
	}

	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		logger.Errorf("could not update application: %v", err)
		return
	}
}

func (s *Service) handleQueueEvent(ctx context.Context, ev *si.EventRecord) {
	logger := log.FromContext(ctx)
	logger.Debugf("adding queue event to accumulator: %v", ev)

	var daoQueue dao.PartitionQueueDAOInfo
	if err := json.Unmarshal([]byte(ev.GetState()), &daoQueue); err != nil {
		logger.Errorw("Failed to unmarshal queue state from event", "error", err)
		return
	}

	isNew := ev.GetEventChangeType() == si.EventRecord_ADD &&
		(ev.GetEventChangeDetail() == si.EventRecord_DETAILS_NONE || ev.GetEventChangeDetail() == si.EventRecord_QUEUE_DYNAMIC)

	var queue *model.Queue
	if isNew {
		s.partitionAccumulator.add(ev)
		queue = &model.Queue{
			Metadata: model.Metadata{
				CreatedAtNano: ev.TimestampNano,
			},
			PartitionQueueDAOInfo: daoQueue,
		}

		if err := s.repo.InsertQueue(ctx, queue); err != nil {
			logger.Errorf("could not insert queue: %v", err)
			return
		}

		return
	}

	queue, err := s.repo.GetQueueInPartition(ctx, daoQueue.PartitionID, daoQueue.QueueName)
	if err != nil {
		logger.Errorf("could not get queue by partition name and queue name: %v", err)
		return
	}

	queue.MergeFrom(&daoQueue)
	if ev.GetEventChangeType() == si.EventRecord_REMOVE {
		queue.DeletedAtNano = &ev.TimestampNano
	}

	if err := s.repo.UpdateQueue(ctx, queue); err != nil {
		logger.Errorf("could not update queue: %v", err)
		return
	}
}
