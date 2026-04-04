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

// health
type HealthResponse struct {
	Status    ReadyStatus `json:"status"`
	Version   string      `json:"version"`
	Timestamp time.Time   `json:"timestamp"`
	Checks    CheckList   `json:"checks"`
}

type CheckList struct {
	Database ReadyStatus `json:"database"`
}
