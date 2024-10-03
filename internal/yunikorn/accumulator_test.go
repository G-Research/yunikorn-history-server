package yunikorn

import (
	"context"
	"testing"
	"time"

	"sync/atomic"

	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
	"github.com/stretchr/testify/require"
)

func TestAccumulatorCallback(t *testing.T) {
	var callCount atomic.Int32
	callback := func(ctx context.Context, events []*si.EventRecord) {
		t.Logf("callback called with %d events", len(events))
		callCount.Add(1)
	}
	interval := 250 * time.Millisecond
	wait := 300 * time.Millisecond

	acc := newAccumulator(callback, interval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		require.NoError(t, acc.run(ctx))
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
