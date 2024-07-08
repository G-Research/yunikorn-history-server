package health

import (
	"context"
	"sync"
	"time"
)

const defaultTimeout = 3 * time.Second

type Interface interface {
	Liveness(ctx context.Context) *LivenessStatus
	Readiness(ctx context.Context) *ReadinessStatus
}

type Service struct {
	version    string
	startedAt  time.Time
	components []Component
}

func New(version string, components ...Component) *Service {
	return &Service{startedAt: time.Now(), version: version, components: components}
}

// Liveness returns the liveness status of the application indicating if the application is running.
func (h *Service) Liveness(ctx context.Context) *LivenessStatus {
	return NewLivenessStatus(h.startedAt, h.version)
}

// NewLivenessStatus creates a new liveness status report.
func NewLivenessStatus(startedAt time.Time, version string) *LivenessStatus {
	return &LivenessStatus{Common: NewCommon(startedAt, version), Healthy: true}
}

// Readiness returns the readiness status of the application indicating if the application is ready to serve requests.
func (h *Service) Readiness(ctx context.Context) *ReadinessStatus {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(h.components))
	componentStatuses := make([]*ComponentStatus, 0, len(h.components))
	for _, component := range h.components {
		go func(component Component) {
			defer wg.Done()
			componentStatus := component.Check(ctx)
			mutex.Lock()
			componentStatuses = append(componentStatuses, componentStatus)
			mutex.Unlock()
		}(component)
	}
	wg.Wait()
	return NewReadinessStatus(h.startedAt, h.version, componentStatuses)
}

// NewReadinessStatus creates a new aggregated readiness status report.
func NewReadinessStatus(startedAt time.Time, version string, componentStatuses []*ComponentStatus) *ReadinessStatus {
	healthy := true
	for _, componentStatus := range componentStatuses {
		if !componentStatus.Healthy {
			healthy = false
			break
		}
	}
	return &ReadinessStatus{
		Common:            NewCommon(startedAt, version),
		Healthy:           healthy,
		ComponentStatuses: componentStatuses,
	}
}
