package domain

import (
	"context"
)

type CoverStatus string

// Размеры обложки
const (
	CoverWidth  = 400
	CoverHeight = 600
	ThumbWidth  = 100
	ThumbHeight = 150
	JPEGQuality = 85
)

// Статус обложки
const (
	CoverStatusNone       CoverStatus = "none"
	CoverStatusProcessing CoverStatus = "processing"
	CoverStatusReady      CoverStatus = "ready"
	CoverStatusFailed     CoverStatus = "failed"
)

type CoverRepository interface {
	UpdateStatus(ctx context.Context, id string, status CoverStatus, coverPath, thumbPath, errorMsg string) error
}
