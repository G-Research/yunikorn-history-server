package health

import (
	"os"
	"time"
)

// Common is a common health status report contained in both liveness and readiness reports.
type Common struct {
	Host      string        `json:"host"`
	StartedAt time.Time     `json:"startedAt"`
	Uptime    time.Duration `json:"uptime"`
	Version   string        `json:"version"`
}

func NewCommon(startedAt time.Time, version string) Common {
	hostname, _ := os.Hostname()
	return Common{Host: hostname, StartedAt: startedAt, Uptime: time.Since(startedAt), Version: version}
}

type LivenessStatus struct {
	Common
	Healthy bool `json:"healthy"`
}

type ReadinessStatus struct {
	Common
	Healthy           bool               `json:"healthy"`
	ComponentStatuses []*ComponentStatus `json:"componentStatuses"`
}
