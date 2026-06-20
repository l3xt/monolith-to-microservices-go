package s3

import (
	"context"
	"fmt"
	"io"
)

type StorageProvider interface {
	Get(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error)
	Upload(ctx context.Context, bucketName, objectPath string, reader io.Reader, size int64, contentType string) error
	GetURL(bucketName, objectName string) (string, error)
}

type ImageStorage struct {
	provider   StorageProvider
	bucketName string
}

func NewImageStorage(provider StorageProvider, bucket string) *ImageStorage {
	return &ImageStorage{
		provider:   provider,
		bucketName: bucket,
	}
}

func (s *ImageStorage) GetImage(ctx context.Context, path string) (io.ReadCloser, error) {
	file, err := s.provider.Get(ctx, s.bucketName, path)
	if err != nil {
		return nil, fmt.Errorf("ImageStorage.GetImage: %w", err)
	}
	return file, nil
}

func (s *ImageStorage) UploadImage(ctx context.Context, path string, file io.Reader, size int64, contentType string) error {
	err := s.provider.Upload(ctx, s.bucketName, path, file, size, contentType)
	if err != nil {
		return fmt.Errorf("ImageStorage.UploadImage: %w", err)
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
