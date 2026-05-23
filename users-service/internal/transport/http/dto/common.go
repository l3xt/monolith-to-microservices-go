package dto

import "time"

type ReadyStatus string

const (
	StatusReady ReadyStatus = "ok"
	StatusError ReadyStatus = "error"
)

// Ошибка для API
type ErrorResponse struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []ErrorDetail `json:"details"`
	RequestID string        `json:"request_id"`
}

// Конкретика ошибки
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Health
type HealthResponse struct {
	Status    ReadyStatus      `json:"status"`
	Service   string           `json:"service"`
	Version   string           `json:"version"`
	Checks    map[string]Check `json:"checks"`
	Timestamp time.Time        `json:"timestamp"`
}

type Check struct {
	Status   ReadyStatus `json:"status"`
	Duration string      `json:"duration"`
	Error    string      `json:"error,omitempty"`
}
