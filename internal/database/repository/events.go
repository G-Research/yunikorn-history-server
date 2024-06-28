package repository

import (
	"context"
	"sync"

	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"

	"github.com/G-Research/yunikorn-history-server/internal/yunikorn/model"
)

type EventRepository interface {
	// Counts returns a map of event types to their counts.
	Counts(ctx context.Context) (model.EventTypeCounts, error)
	// Record increments the count of the given event type.
	Record(ctx context.Context, event *si.EventRecord) error
}

// InMemoryEventRepository is an in-memory implementation of the EventRepository interface.
// TODO: This implementation is not resilient to crashes and will lose all data when the process is restarted.
type InMemoryEventRepository struct {
	mutex  sync.Mutex
	counts model.EventTypeCounts
}

func NewInMemoryEventRepository() *InMemoryEventRepository {
	return &InMemoryEventRepository{
		counts: make(model.EventTypeCounts),
	}
}

func (r *InMemoryEventRepository) Counts(ctx context.Context) (model.EventTypeCounts, error) {
	return r.counts, nil
}

func (r *InMemoryEventRepository) Record(ctx context.Context, event *si.EventRecord) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	key := model.EventTypeKey{Type: event.GetType(), ChangeType: event.GetEventChangeType()}
	if count, exists := r.counts[key]; exists {
		r.counts[key] = count + 1
	} else {
		r.counts[key] = 1
	}
	return nil
}
