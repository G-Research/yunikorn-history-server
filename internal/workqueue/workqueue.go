package workqueue

import (
	"context"
	"errors"
	"github.com/G-Research/yunikorn-history-server/internal/log"
	"github.com/google/uuid"
	"sync"
	"sync/atomic"
	"time"
)

var ErrNotStarted = errors.New("workqueue not started")

const maxBackoff = 5 * time.Minute
const maxRetries = 20
const initialDelay = 1 * time.Second

type Option func(*WorkQueue)

// WithName sets the name of the workqueue.
func WithName(name string) Option {
	return func(wq *WorkQueue) {
		wq.name = name
	}
}

// WithInitialDelay sets the initial exponential backoff delay of the workqueue.
func WithInitialDelay(delay time.Duration) Option {
	return func(wq *WorkQueue) {
		wq.initialDelay = delay
	}
}

func WithGracefulShutdown(gracePeriod time.Duration) Option {
	return func(wq *WorkQueue) {
		wq.gracePeriod = gracePeriod
	}
}

type JobOption func(*item)

// WithJobName sets the name of the job.
func WithJobName(name string) JobOption {
	return func(i *item) {
		i.name = name
	}
}

type Job func(context.Context) error

type item struct {
	id   uuid.UUID
	name string
	job  Job
}

func newItem(job Job, opts ...JobOption) *item {
	i := &item{
		id:  uuid.New(),
		job: job,
	}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

type WorkQueue struct {
	name         string
	mutex        sync.Mutex
	signal       chan struct{}
	queue        []*item
	initialDelay time.Duration
	started      bool
	// gracePeriod is the graceful shutdown period to wait for jobs to finish before shutting down the workqueue.
	gracePeriod time.Duration
	// running is the number of jobs currently running.
	running int32
	cancel  context.CancelFunc
}

func NewWorkQueue(opts ...Option) *WorkQueue {
	wq := &WorkQueue{
		signal:       make(chan struct{}),
		queue:        make([]*item, 0),
		initialDelay: initialDelay,
	}
	for _, opt := range opts {
		opt(wq)
	}
	return wq
}

// Add adds a job to the workqueue.
func (w *WorkQueue) Add(job Job, opts ...JobOption) error {
	if !w.started {
		return ErrNotStarted
	}
	w.mutex.Lock()
	w.queue = append(w.queue, newItem(job, opts...))
	w.mutex.Unlock()
	w.signal <- struct{}{}
	return nil
}

// pop removes and returns the first job from the workqueue.
func (w *WorkQueue) pop() *item {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if len(w.queue) == 0 {
		return nil
	}

	f := w.queue[0]
	w.queue = w.queue[1:]
	return f
}

// Run starts the workqueue which processes jobs and retries them on failure with exponential backoff.
func (w *WorkQueue) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger = logger.With("component", "workqueue")
	if w.name != "" {
		logger = logger.With("workqueue", w.name)
	}
	ctx = log.ToContext(ctx, logger)

	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	defer cancel()

	jobCtx, jobCtxCancel := context.WithCancel(context.Background())
	defer jobCtxCancel()

	logger.Info("workqueue starting")

	w.started = true

	for {
		select {
		case <-ctx.Done():
			logger.Warnw("workqueue shutting down")
			running := atomic.LoadInt32(&w.running)
			if w.gracePeriod > 0 && running > 0 {
				logger.Infow(
					"workqueue still has pending jobs, waiting for grace period to expire...",
					"gracePeriod", w.gracePeriod.String(), "runningJobs", running,
				)
				<-time.After(w.gracePeriod)
			}
			w.Shutdown()
			return ctx.Err()
		case <-w.signal:
			elem := w.pop()
			if elem == nil {
				logger.Warnw("workqueue received a signal but queue is empty")
				continue
			}
			go w.executeWithRetry(jobCtx, elem)
		}
	}
}

// Started returns true if the workqueue is started.
func (w *WorkQueue) Started() bool {
	return w.started
}

// Shutdown stops the workqueue.
func (w *WorkQueue) Shutdown() {
	if w.started {
		w.started = false
		w.cancel()
		close(w.signal)
	}
}

// executeWithRetry executes a job and retries it on failure with exponential backoff.
func (w *WorkQueue) executeWithRetry(ctx context.Context, item *item) {
	logger := log.FromContext(ctx)
	logger = logger.With("jobId", item.id)
	if item.name != "" {
		logger = logger.With("jobName", item.name)
	}

	atomic.AddInt32(&w.running, 1)
	defer atomic.AddInt32(&w.running, -1)

	now := time.Now()

	backoff := time.Duration(0)
	retries := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
			if retries > maxRetries {
				logger.Errorw("workqueue reached max retries for job", "retryCount", retries)
			}
			err := item.job(ctx)
			if err == nil {
				logger.Infow("workqueue successfully processed job", "duration", time.Since(now).String())
				return
			}
			if backoff == 0 {
				backoff = w.initialDelay
			} else if backoff < maxBackoff {
				backoff *= 2
			}
			if err != nil {
				logger.Errorw(
					"workqueue failed to process job, retrying after exponential backoff...",
					"error", err, "retryCount", retries, "backoff", backoff.String(),
				)
			}
			retries++
		}
	}
}
