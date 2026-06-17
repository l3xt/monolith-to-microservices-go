package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type CoverStatus string

var (
	ErrCoverNotFound     = errors.New("cover not found")
	ErrInvalidCoverFile  = errors.New("invalid cover file")
	ErrCoverExceededSize = errors.New("cover exceeded max size")
	ErrInvalidCoverType  = errors.New("invalid cover type")
)

const (
	CoverStatusNone       CoverStatus = "none"
	CoverStatusProcessing CoverStatus = "processing"
	CoverStatusReady      CoverStatus = "ready"
	CoverStatusFailed     CoverStatus = "failed"
)

type Cover struct {
	ID           uuid.UUID
	BookID       uuid.UUID
	Status       CoverStatus
	OriginalPath *string
	CoverPath    *string
	ThumbPath    *string
	CoverURL     string
	ThumbURL     string
	Error        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}

type CoverRepository interface {
	Create(ctx context.Context, cover *Cover) error
	GetByBookID(ctx context.Context, bookID uuid.UUID) (*Cover, error)
	UpdateStatus(ctx context.Context, id string, status CoverStatus, coverPath, thumbPath, errorMsg string) error
	DeleteByBookID(ctx context.Context, bookID uuid.UUID) error
}
