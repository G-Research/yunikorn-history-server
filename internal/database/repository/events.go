package repository

import (
	"context"
	"fmt"
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
	// We must lock and make a copy of the original map to avoid
	// "concurrent map read and map write" panics, if the caller
	// of this func reads from the returned result of this func.
	r.mutex.Lock()
	defer r.mutex.Unlock()
	countsCopy := model.EventTypeCounts{}
	for k, v := range r.counts {
		countsCopy[k] = v
	}

	return countsCopy, nil
}

func (r *InMemoryEventRepository) Record(ctx context.Context, event *si.EventRecord) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	key := getKey(event)
	if count, exists := r.counts[key]; exists {
		r.counts[key] = count + 1
	} else {
		r.counts[key] = 1
	}
	return nil
}

// getKey returns a key for the given event record which is a combination of the event type and the change type.
func getKey(e *si.EventRecord) string {
	return fmt.Sprintf("%s-%s", e.GetType().String(), e.GetEventChangeType().String())
}
