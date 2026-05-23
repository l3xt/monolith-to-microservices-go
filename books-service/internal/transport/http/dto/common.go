package dto

import "time"

const minLimit = 10

type ReadyStatus string

const (
	StatusReady ReadyStatus = "ok"
	StatusError ReadyStatus = "error"
)

// Метаданные пагинации
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

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

func NewPagination(page, limit, total int) *Pagination {
	if limit <= 0 {
		limit = minLimit
	}
	if page <= 0 {
		page = 1
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return &Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
