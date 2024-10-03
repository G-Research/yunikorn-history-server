package yunikorn

import (
	"context"
	"testing"
	"time"

	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func TestAccumulatorCallback(t *testing.T) {
	callCount := atomic.NewInt32(0)
	callback := func(ctx context.Context, events []*si.EventRecord) {
		t.Logf("callback called with %d events", len(events))
		callCount.Inc()
	}
	interval := 250 * time.Millisecond
	wait := 300 * time.Millisecond

	acc := newAccumulator(callback, interval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		acc.run(ctx)
	}()

	time.Sleep(wait)
	require.Equal(t, int32(0), callCount.Load())

	acc.add(&si.EventRecord{})
	time.Sleep(wait)
	require.Equal(t, int32(1), callCount.Load())

	acc.add(&si.EventRecord{})
	acc.add(&si.EventRecord{})
	acc.add(&si.EventRecord{})
	time.Sleep(wait)
	require.Equal(t, int32(2), callCount.Load())
}
