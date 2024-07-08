package workqueue

import (
	"context"
	"errors"
	"github.com/G-Research/yunikorn-history-server/test/log"
	"github.com/google/uuid"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkQueue(t *testing.T) {
	wq := NewWorkQueue(WithInitialDelay(time.Second))
	assert.NotNil(t, wq)
	assert.Empty(t, wq.queue)
}

func TestAdd(t *testing.T) {
	t.Run("add job to queue when workqueue is not started", func(t *testing.T) {
		wq := NewWorkQueue(WithInitialDelay(time.Second))
		job := func(ctx context.Context) error { return nil }

		assert.ErrorIs(t, wq.Add(job), ErrNotStarted)
		assert.Len(t, wq.queue, 0)
	})
	t.Run("add job to queue when workqueue is started", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		wq := NewWorkQueue(WithInitialDelay(time.Second))
		processed := false
		job := func(ctx context.Context) error { processed = true; return nil }

		go wq.Run(ctx)

		assert.Eventually(t, func() bool {
			return wq.started
		}, 250*time.Millisecond, 50*time.Millisecond)

		assert.NoError(t, wq.Add(job))
		assert.Eventually(t, func() bool {
			return processed
		}, 250*time.Millisecond, 50*time.Millisecond)
	})
}

func TestPop(t *testing.T) {
	expectedJobID := uuid.New()

	tests := []struct {
		name          string
		initialQueue  []*item
		expectedJobID uuid.UUID
		expectedNil   bool
		expectedLen   int
	}{
		{
			name: "Non-empty queue",
			initialQueue: []*item{
				{id: expectedJobID, job: func(ctx context.Context) error { return nil }},
			},
			expectedJobID: expectedJobID,
			expectedLen:   0,
		},
		{
			name:          "Empty queue",
			initialQueue:  []*item{},
			expectedJobID: uuid.Nil,
			expectedNil:   true,
			expectedLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wq := WorkQueue{}
			wq.queue = tt.initialQueue

			retJob := wq.pop()
			if tt.expectedNil {
				assert.Nil(t, retJob)
			} else {
				assert.NotNil(t, retJob)
				assert.Equal(t, tt.expectedJobID, retJob.id)
			}
			assert.Len(t, wq.queue, tt.expectedLen)
		})
	}
}

func TestRunAndExecuteWithRetry(t *testing.T) {
	ctx, _ := log.GetTestLogger(context.Background())

	ctx, cancel := context.WithTimeout(ctx, 1000*time.Second)
	defer cancel()

	wq := NewWorkQueue(WithInitialDelay(50 * time.Millisecond))

	go wq.Run(ctx)

	assert.Eventually(t, func() bool {
		return wq.started
	}, 250*time.Millisecond, 50*time.Millisecond)

	jobRunCount := int32(0)
	job := func(ctx context.Context) error {
		atomic.AddInt32(&jobRunCount, 1)
		if atomic.LoadInt32(&jobRunCount) < 3 {
			return errors.New("retry")
		}
		return nil
	}

	assert.NoError(t, wq.Add(job))

	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&jobRunCount) == 3
	}, 1000*time.Millisecond, 50*time.Millisecond)
}

func TestRun_Shutdown(t *testing.T) {
	ctx, _ := log.GetTestLogger(context.Background())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wq := NewWorkQueue(WithInitialDelay(50 * time.Millisecond))

	go wq.Run(ctx)

	assert.Eventually(t, func() bool {
		return wq.started
	}, 250*time.Millisecond, 50*time.Millisecond)

	wq.Shutdown()

	assert.Eventually(t, func() bool {
		return !wq.started
	}, 250*time.Millisecond, 50*time.Millisecond)

	// Add a job after shutdown
	job := func(ctx context.Context) error { return nil }
	assert.ErrorIs(t, wq.Add(job), ErrNotStarted)

	assert.Len(t, wq.queue, 0)
}

func TestJobExecutionStopsOnContextCancel(t *testing.T) {
	ctx, _ := log.GetTestLogger(context.Background())

	ctx, cancel := context.WithCancel(ctx)

	wq := NewWorkQueue(WithName("test"), WithInitialDelay(20*time.Millisecond))

	jobRunCount := int32(0)
	job := func(ctx context.Context) error {
		atomic.AddInt32(&jobRunCount, 1)
		if atomic.LoadInt32(&jobRunCount) == 2 {
			cancel()
		}
		return errors.New("some error")
	}

	go wq.Run(ctx)

	assert.Eventually(t, func() bool {
		return wq.started
	}, 250*time.Millisecond, 50*time.Millisecond)

	assert.NoError(t, wq.Add(job))

	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&jobRunCount) == 2
	}, 100*time.Millisecond, 20*time.Millisecond)
	assert.Never(t, func() bool {
		return atomic.LoadInt32(&jobRunCount) > 2
	}, 100*time.Millisecond, 20*time.Millisecond)
}

func TestJobExecutionStopsOnContextCancelWithGracefulShutdown(t *testing.T) {
	ctx, _ := log.GetTestLogger(context.Background())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wq := NewWorkQueue(WithName("test"), WithInitialDelay(20*time.Millisecond), WithGracefulShutdown(200*time.Millisecond))

	job := func(ctx context.Context) error {
		assert.Equal(t, wq.running, int32(1))
		assert.Eventually(t, func() bool {
			return !wq.started && wq.running == 1
		}, 400*time.Millisecond, 50*time.Millisecond)
		return nil
	}

	go wq.Run(ctx)

	assert.Eventually(t, func() bool {
		return wq.started
	}, 250*time.Millisecond, 50*time.Millisecond)

	assert.NoError(t, wq.Add(job))

	cancel()

	assert.Eventually(t, func() bool {
		return wq.running == 0
	}, 600*time.Millisecond, 50*time.Millisecond)
}

func TestJobExecutionStopsOnShutdownWithGracefulShutdown(t *testing.T) {
	ctx, _ := log.GetTestLogger(context.Background())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wq := NewWorkQueue(WithName("test"), WithInitialDelay(20*time.Millisecond), WithGracefulShutdown(200*time.Millisecond))

	job := func(ctx context.Context) error {
		assert.Equal(t, wq.running, int32(1))
		assert.Eventually(t, func() bool {
			return !wq.started
		}, 250*time.Millisecond, 50*time.Millisecond)
		return nil
	}

	go wq.Run(ctx)

	assert.Eventually(t, func() bool {
		return wq.started
	}, 250*time.Millisecond, 50*time.Millisecond)

	assert.NoError(t, wq.Add(job))

	wq.Shutdown()

	assert.Eventually(t, func() bool {
		return wq.running == 0
	}, 500*time.Millisecond, 50*time.Millisecond)
}
