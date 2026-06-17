package dto

import (
	"bookshelf/books-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Загрузка обложки
type CoverUploadResponse struct {
	CoverID uuid.UUID          `json:"cover_id"`
	Status  domain.CoverStatus `json:"status"`
	Message string             `json:"message"`
}

// Обложка
type CoverResponse struct {
	Status   domain.CoverStatus `json:"status"`
	CoverURL *string            `json:"cover_url,omitempty"`
	ThumbURL *string            `json:"thumb_url,omitempty"`
	Message  *string            `json:"message,omitempty"`
}

// Детальный статус
type CoverStatusResponse struct {
	CoverID     string     `json:"cover_id"`
	CoverURL    *string    `json:"cover_url,omitempty"`
	ThumbURL    *string    `json:"thumb_url,omitempty"`
	Error       *string    `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
