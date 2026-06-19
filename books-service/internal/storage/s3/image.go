package s3

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/pkg/utils"
	"context"
	"fmt"
	"io"
)

type StorageProvider interface {
	Upload(ctx context.Context, bucketName, objectPath string, reader io.Reader, size int64, contentType string) error
}

type ImageStorage struct {
	provider   StorageProvider
	bucketName string
}

func NewImageStorage(provider StorageProvider, bucket string) *ImageStorage {
	return &ImageStorage{
		provider: provider,
		bucketName: bucket,
	}
}

func (s *ImageStorage) UploadImage(ctx context.Context, path string, file io.Reader, size int64) error {
	contentType, file, err := utils.DetectContentType(file)
	if err != nil {
		return fmt.Errorf("FileStorage UploadImage: detect content type: %w", err)
	}

	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" && contentType != "image/jpg" {
		return fmt.Errorf("FileStorage UploadImage: content type: %w", domain.ErrInvalidCoverType)
	}

	return s.provider.Upload(ctx, s.bucketName, path, file, size, contentType)
}
