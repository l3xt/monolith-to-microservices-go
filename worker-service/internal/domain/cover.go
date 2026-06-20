package domain

import (
	"context"
	"errors"
)

type CoverStatus string

var (
	ErrCoverNotFound     = errors.New("cover not found")
	ErrInvalidCoverFile  = errors.New("invalid cover file")
	ErrCoverExceededSize = errors.New("cover exceeded max size")
	ErrInvalidCoverType  = errors.New("invalid cover type")
)

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
