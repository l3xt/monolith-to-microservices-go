package minio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	ErrBucketNotExists = errors.New("bucket not exists")
)

type Config struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
}

type Storage struct {
	client         *minio.Client
	publicEndpoint string
}

func New(cfg Config) (*Storage, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("Storage New: failed to initialize minio client: %w", err)
	}
	return &Storage{
		client:         minioClient,
		publicEndpoint: cfg.PublicEndpoint,
	}, nil
}

func (s *Storage) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("Storage EnsureBucket: failed to check if bucket exists: %w", err)
	}
	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("Storage EnsureBucket: failed to create bucket: %w", err)
		}
	}
	return nil
}

func (s *Storage) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := s.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("Storage Upload: failed to upload object: %w", err)
	}
	return nil
}

func (s *Storage) GetURL(bucketName, objectName string) (string, error) {
	return url.JoinPath(s.publicEndpoint, bucketName, objectName)
}

func (s *Storage) GetPresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	// Временная ссылка
	u, err := s.client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *Storage) Delete(ctx context.Context, bucketName, objectName string) error {
	err := s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("Storage Delete: failed to remove object: %w", err)
	}
	return nil
}

func (s *Storage) HealthCheck(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("Storage HealthCheck: failed to check if bucket exists: %w", err)
	}
	if !exists {
		return ErrBucketNotExists
	}
	return nil
}
