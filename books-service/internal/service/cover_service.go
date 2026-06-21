package service

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/pkg/utils"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
)

const (
	maxImageSize = 10 * 1024 * 1024 // 10 МБ
)

var (
	ErrInvalidCoverStatus = errors.New("invalid cover status")
)

type EventPublisher interface {
	PublishImageCompressEvent(ctx context.Context, bookID, coverID uuid.UUID, originalPath string) error
}

type ImageStorage interface {
	UploadImage(ctx context.Context, path string, reader io.Reader, size int64) error
	Delete(ctx context.Context, path string) error
	GetURL(objectPath string) (string, error)
}

type CoverService struct {
	bookRepo  BookRepository
	coverRepo domain.CoverRepository
	storage   ImageStorage
	publisher EventPublisher
}

func NewCoverService(bookRepo BookRepository, coverRepo domain.CoverRepository, storage ImageStorage, publisher EventPublisher) *CoverService {
	return &CoverService{
		bookRepo:  bookRepo,
		coverRepo: coverRepo,
		storage:   storage,
		publisher: publisher,
	}
}

func (s *CoverService) UploadBookCover(ctx context.Context, userID, bookID uuid.UUID, reader io.Reader, size int64) (uuid.UUID, error) {
	// Проверка книги
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CoverService.UploadBookCover: get book: %w", err)
	}

	if book.UserID != userID {
		return uuid.Nil, domain.ErrNotBookOwner
	}
	if size <= 0 {
		return uuid.Nil, fmt.Errorf("CoverService.UploadBookCover: size is zero or negative: %w", domain.ErrInvalidCoverFile)
	}
	if size > maxImageSize {
		return uuid.Nil, fmt.Errorf("CoverService.UploadBookCover: size exceed max: %w", domain.ErrCoverExceededSize)
	}

	contentType, reader, err := utils.DetectContentType(reader)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CoverService.UploadBookCover: detect content type: %w", err)
	}

	ext := utils.MapContentTypeToExt(contentType)
	if ext == "" {
		return uuid.Nil, fmt.Errorf("CoverService.UploadBookCover: %w", domain.ErrInvalidCoverFile)
	}

	originalPath := fmt.Sprintf("covers/%s/original%s", bookID.String(), ext)

	if err := s.storage.UploadImage(ctx, originalPath, reader, size); err != nil {
		return uuid.Nil, fmt.Errorf("CoverService.UploadBookCover: upload cover: %w", err)
	}

	cover := &domain.Cover{
		BookID:       bookID,
		Status:       domain.CoverStatusProcessing,
		OriginalPath: &originalPath,
	}

	if err := s.coverRepo.Create(ctx, cover); err != nil {
		return uuid.Nil, fmt.Errorf("CoverService.UploadBookCover: create cover: %w", err)
	}

	if err := s.publisher.PublishImageCompressEvent(ctx, bookID, cover.ID, originalPath); err != nil {
		return uuid.Nil, fmt.Errorf("CoverService.UploadBookCover: publish event: %w", err)
	}

	return cover.ID, nil
}

func (s *CoverService) GetCover(ctx context.Context, bookID uuid.UUID) (*domain.Cover, error) {
	cover, err := s.coverRepo.GetByBookID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("CoverService.GetCover: get cover: %w", err)
	}

	if cover.Status == domain.CoverStatusReady {
		if cover.CoverPath != nil && cover.CoverURL == nil {
			coverURL, err := s.storage.GetURL(*cover.CoverPath)
			if err != nil {
				return nil, fmt.Errorf("CoverService.GetCover: get cover: %w", err)
			}
			cover.CoverURL = &coverURL
		}

		if cover.ThumbPath != nil && cover.ThumbURL == nil {
			thumbURL, err := s.storage.GetURL(*cover.ThumbPath)
			if err != nil {
				return nil, fmt.Errorf("CoverService.GetCover: get cover: %w", err)
			}
			cover.ThumbURL = &thumbURL
		}
	}

	return cover, nil
}

func (s *CoverService) DeleteCover(ctx context.Context, userID, bookID uuid.UUID) error {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return fmt.Errorf("CoverService.DeleteCover: get book: %w", err)
	}

	if book.UserID != userID {
		return fmt.Errorf("CoverService.DeleteCover: %w", domain.ErrNotBookOwner)
	}

	cover, err := s.coverRepo.GetByBookID(ctx, bookID)
	if err != nil {
		if errors.Is(err, domain.ErrCoverNotFound) {
			return nil // обложки и так нет
		}
		return fmt.Errorf("CoverService.DeleteCover: get cover: %w", err)
	}

	if cover.OriginalPath != nil {
		if err := s.storage.Delete(ctx, *cover.OriginalPath); err != nil {
			return fmt.Errorf("CoverService.DeleteCover: delete cover: %w", err)
		}
	}

	if cover.CoverPath != nil {
		if err := s.storage.Delete(ctx, *cover.CoverPath); err != nil {
			return fmt.Errorf("CoverService.DeleteCover: delete cover: %w", err)
		}
	}

	if cover.ThumbPath != nil {
		if err := s.storage.Delete(ctx, *cover.ThumbPath); err != nil {
			return fmt.Errorf("CoverService.DeleteCover: delete cover: %w", err)
		}
	}

	if err := s.coverRepo.DeleteByBookID(ctx, bookID); err != nil {
		return fmt.Errorf("CoverService.DeleteCover: delete cover: %w", err)
	}

	if err := s.bookRepo.UpdateCover(ctx, bookID, domain.CoverStatusNone, nil, nil); err != nil {
		return fmt.Errorf("CoverService.DeleteCover: update cover: %w", err)
	}

	return nil
}
