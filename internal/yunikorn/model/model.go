package model

import (
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
)

// EventTypeCounts is a map of event types to their counts.
type EventTypeCounts map[string]int

// EventTypeKey is a key for the EventTypeCounts map and is a combination of the event type and the change type.
type EventTypeKey struct {
	Type       si.EventRecord_Type
	ChangeType si.EventRecord_ChangeType
}
