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
	Delete(ctx context.Context, bucketName, objectName string) error
	GetURL(bucketName, objectName string) (string, error)
	HealthCheck(ctx context.Context, bucketName string) error
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

	if err := s.provider.Upload(ctx, s.bucketName, path, file, size, contentType); err != nil {
		return fmt.Errorf("ImageStorage UploadImage: upload object: %w", err)
	}

	return nil
}

func (s *ImageStorage) Delete(ctx context.Context, path string) error {
	if err := s.provider.Delete(ctx, s.bucketName, path); err != nil {
		return fmt.Errorf("ImageStorage Delete: delete object: %w", err)
	}

	return nil
}

func (s *ImageStorage) GetURL(objectPath string) (string, error) {
	url, err := s.provider.GetURL(s.bucketName, objectPath)
	if err != nil {
		return "", fmt.Errorf("ImageStorage.GetURL: %w", err)
	}
	return url, nil
}

func (s *ImageStorage) HealthCheck(ctx context.Context) error {
	if err := s.provider.HealthCheck(ctx, s.bucketName); err != nil {
		return fmt.Errorf("ImageStorage HealthCheck: check bucket: %w", err)
	}
	return nil
}
