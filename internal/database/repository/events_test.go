package repository

import (
	"context"
	"testing"

	"github.com/G-Research/yunikorn-scheduler-interface/lib/go/si"
	"github.com/stretchr/testify/assert"
)

func TestGetKey(t *testing.T) {
	tests := []struct {
		name     string
		event    *si.EventRecord
		expected string
	}{
		{
			name: "Event Type APP, Change Type SET",
			event: &si.EventRecord{
				Type:            si.EventRecord_APP,
				EventChangeType: si.EventRecord_SET,
			},
			expected: "APP-SET",
		},
		{
			name: "Event Type Delete, Change Type None",
			event: &si.EventRecord{
				Type:            si.EventRecord_NODE,
				EventChangeType: si.EventRecord_ADD,
			},
			expected: "NODE-ADD",
		},
		{
			name: "Event Type Update, Change Type Delete",
			event: &si.EventRecord{
				Type:            si.EventRecord_QUEUE,
				EventChangeType: si.EventRecord_REMOVE,
			},
			expected: "QUEUE-REMOVE",
		},
		{
			name:     "Event Type None, Change Type None",
			event:    &si.EventRecord{},
			expected: "UNKNOWN_EVENTRECORD_TYPE-NONE",
		},
		{
			name: "Event Type App, Change Type None",
			event: &si.EventRecord{
				Type: si.EventRecord_APP,
			},
			expected: "APP-NONE",
		},
		{
			name: "Event Type None, Change Type Add",
			event: &si.EventRecord{
				EventChangeType: si.EventRecord_ADD,
			},
			expected: "UNKNOWN_EVENTRECORD_TYPE-ADD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getKey(tt.event)
			if result != tt.expected {
				t.Errorf("got %s, expected %v", result, tt.expected)
			}
		})
	}
}

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
	repository.counts[getKey(event)] = 1

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
	assert.Equal(t, 2, repository.counts[getKey(event1)])
	assert.Equal(t, 1, repository.counts[getKey(event3)])
}
