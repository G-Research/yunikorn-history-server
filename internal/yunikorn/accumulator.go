package yunikorn

import (
	"context"
	"sync"
	"time"

	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
)

// accumulator is a struct that accumulates events and processes them
// based on a timer.
type accumulator struct {
	// events is a channel that accumulates events.
	// there should only be one event slice in the channel at any time.
	// if the channel is empty, something is working on events.
	events chan []*si.EventRecord
	// timerNotifier notifies the timer manager goroutine to do an action
	// on the timer based on the state of events.
	// This notification should be raised when either an event is added,
	// or events are drained.
	timerNotifier chan struct{}
	fn            func(ctx context.Context, events []*si.EventRecord)
	idleInterval  time.Duration
}

func newAccumulator(fn func(ctx context.Context, events []*si.EventRecord), idleInterval time.Duration) *accumulator {
	acc := &accumulator{
		events:        make(chan []*si.EventRecord, 1),
		timerNotifier: make(chan struct{}, 1),
		fn:            fn,
		idleInterval:  idleInterval,
	}
	acc.events <- []*si.EventRecord{}
	return acc
}

// add is responsible for adding an event to the accumulator.
// When an event is added, the timer should be restarted.
func (a *accumulator) add(event *si.EventRecord) {
	events := <-a.events
	a.events <- append(events, event)
	select {
	case a.timerNotifier <- struct{}{}:
	default:
	}
}

// run is responsible for running the accumulator.
func (a *accumulator) run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(2)

	// initial state of the timer should be stopped
	timer := time.NewTicker(a.idleInterval)
	timer.Stop()

	// listen for timer instructions
	// this goroutine drains the timerInstructions channel
	// and acts on the instruction
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-a.timerNotifier:
				events := <-a.events
				if len(events) == 0 {
					timer.Stop()
					a.events <- events
					return
				}
				timer.Reset(a.idleInterval)
				a.events <- events
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				events := <-a.events
				a.events <- []*si.EventRecord{}
				// if timer channel is already occupied, no need to push the event to it
				select {
				case a.timerNotifier <- struct{}{}:
				default:
				}

				// TODO: figure out if we want to push it to a worker pool
				a.fn(ctx, events)
			}
		}
	}()

	wg.Wait()

	return nil
}
