package service

import (
	"bookshelf/pkg/utils"
	"bookshelf/worker-service/internal/apperror"
	"bookshelf/worker-service/internal/domain"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
)

var ErrEmptyFileExt = errors.New("empty file extension")

type BookRepository interface {
	UpdateCover(ctx context.Context, bookID uuid.UUID, coverURL, thumbURL *string, status domain.CoverStatus) error
}

type CoverRepository interface {
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CoverStatus, coverPath, thumbPath, errorMsg *string) error
}

type ImageStorage interface {
	GetImage(ctx context.Context, path string) (io.ReadCloser, error)
	UploadImage(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error
	GetURL(objectPath string) (string, error)
}

type ImageProcessor interface {
	CreateCover(in io.Reader) ([]byte, error)
	CreateThumbnail(in io.Reader) ([]byte, error)
}

type CoverService struct {
	bookRepo  BookRepository
	coverRepo CoverRepository
	storage   ImageStorage
	imgProc   ImageProcessor
}

func NewCoverService(bookRepo BookRepository, coverRepo CoverRepository, storage ImageStorage, imgProc ImageProcessor) *CoverService {
	return &CoverService{
		bookRepo:  bookRepo,
		coverRepo: coverRepo,
		storage:   storage,
		imgProc:   imgProc,
	}
}

func (s *CoverService) ProcessBookCover(ctx context.Context, bookID, coverID uuid.UUID, imagePath string) (err error) {
	defer func() {
		if err != nil {
			// Оборачиваем ошибку, если она является временной
			if errors.Is(err, domain.ErrServiceNotResponding) {
				err = apperror.NewRetryable(err)
			}

			// Используем context.WithoutCancel(ctx), чтобы запрос в БД гарантированно выполнился, даже если оригинальный контекст был отменен.
			var errMsg = err.Error()
			safeCtx := context.WithoutCancel(ctx)
			s.coverRepo.UpdateStatus(safeCtx, coverID, domain.CoverStatusFailed, nil, nil, &errMsg)
			s.bookRepo.UpdateCover(safeCtx, bookID, nil, nil, domain.CoverStatusFailed)
		}
	}()

	reader, err := s.storage.GetImage(ctx, imagePath)
	if err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: get image from storage: %w", err)
	}
	defer reader.Close()

	contentType, imgReader, err := utils.DetectContentType(reader)
	if err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: detect content type: %w", err)
	}

	// создаем буффер для thumb версии
	var thumbBuf bytes.Buffer
	teeReader := io.TeeReader(imgReader, &thumbBuf)

	// Создаем Cover (через teeReader, чтобы заполнить thumbBuf)
	coverBytes, err := s.imgProc.CreateCover(teeReader)
	if err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: create cover: %w", err)
	}

	// Создаем Thumb
	thumbBytes, err := s.imgProc.CreateThumbnail(&thumbBuf)
	if err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: create thumb: %w", err)
	}

	// Загружаем обратно в Minio с динамическим расширением
	ext := utils.MapContentTypeToExt(contentType)
	if ext == "" {
		return fmt.Errorf("ImageService.ProcessBookCover: %w", ErrEmptyFileExt)
	}

	coverPath := fmt.Sprintf("covers/%s/cover%s", bookID, ext)
	if err := s.storage.UploadImage(ctx, coverPath, bytes.NewReader(coverBytes), int64(len(coverBytes)), contentType); err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: upload cover: %w", err)
	}

	thumbPath := fmt.Sprintf("covers/%s/thumb%s", bookID, ext)
	if err := s.storage.UploadImage(ctx, thumbPath, bytes.NewReader(thumbBytes), int64(len(thumbBytes)), contentType); err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: upload thumb: %w", err)
	}

	if err := s.coverRepo.UpdateStatus(ctx, coverID, domain.CoverStatusReady, &coverPath, &thumbPath, nil); err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: update cover status: %w", err)
	}

	coverURL, err := s.storage.GetURL(coverPath)
	if err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: get cover url: %w", err)
	}

	thumbURL, err := s.storage.GetURL(thumbPath)
	if err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: get thumb url: %w", err)
	}

	if err := s.bookRepo.UpdateCover(ctx, bookID, &coverURL, &thumbURL, domain.CoverStatusReady); err != nil {
		return fmt.Errorf("ImageService.ProcessBookCover: update book metadata: %w", err)
	}
	return nil
}
