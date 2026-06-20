package dto

import "time"

type ReadyStatus string

const (
	StatusReady ReadyStatus = "ok"
	StatusError ReadyStatus = "error"
)

// Health
type HealthResponse struct {
	Status    ReadyStatus      `json:"status"`
	Service   string           `json:"service"`
	Version   string           `json:"version"`
	Checks    map[string]Check `json:"checks"`
	Timestamp time.Time        `json:"timestamp"`
}

// Ready
type ReadyResponse struct {
	Ready     bool             `json:"ready"`
	Service   string           `json:"service"`
	Checks    map[string]Check `json:"checks"`
	Timestamp time.Time        `json:"timestamp"`
}

type Check struct {
	Status   ReadyStatus `json:"status"`
	Duration string      `json:"duration"`
	Error    string      `json:"error,omitempty"`
}
