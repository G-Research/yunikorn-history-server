package repository

import (
	"context"
	"testing"

	"github.com/G-Research/yunikorn-history-server/internal/yunikorn/model"

	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryEventRepository_Count(t *testing.T) {
	repository := NewInMemoryEventRepository()
	ctx := context.Background()

	counts, err := repository.Counts(ctx)
	assert.NoError(t, err)
	assert.Empty(t, counts)

	// Record an event and check counts
	event := &si.EventRecord{
		Type:            si.EventRecord_APP,
		EventChangeType: si.EventRecord_ADD,
	}
	repository.counts[model.EventTypeKey{Type: event.GetType(), ChangeType: event.GetEventChangeType()}] = 1

	counts, err = repository.Counts(ctx)
	assert.NoError(t, err)
	assert.Len(t, counts, 1)
}

func TestInMemoryEventRepository_Record(t *testing.T) {
	repository := NewInMemoryEventRepository()
	ctx := context.Background()

	// Record multiple events
	event1 := &si.EventRecord{
		Type:            si.EventRecord_NODE,
		EventChangeType: si.EventRecord_ADD,
	}
	event2 := &si.EventRecord{
		Type:            si.EventRecord_NODE,
		EventChangeType: si.EventRecord_ADD,
	}
	event3 := &si.EventRecord{
		Type:            si.EventRecord_APP,
		EventChangeType: si.EventRecord_REMOVE,
	}

	assert.NoError(t, repository.Record(ctx, event1))
	assert.NoError(t, repository.Record(ctx, event2))
	assert.NoError(t, repository.Record(ctx, event3))

	assert.Len(t, repository.counts, 2)

	// Verify the counts of the specific events
	key1 := model.EventTypeKey{Type: si.EventRecord_NODE, ChangeType: si.EventRecord_ADD}
	key2 := model.EventTypeKey{Type: si.EventRecord_APP, ChangeType: si.EventRecord_REMOVE}
	assert.Equal(t, 2, repository.counts[key1])
	assert.Equal(t, 1, repository.counts[key2])
}
