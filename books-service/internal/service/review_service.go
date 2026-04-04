package service

import (
	"bookshelf/books-service/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const DefaultReviewCountLimit = 20

var (
	ErrEmptyReviewID = errors.New("empty id value")
	ErrEmptyUserID   = errors.New("empty user id value")
)

type ReviewService struct {
	reviewRepo domain.ReviewRepository
	bookRepo   domain.BookRepository
}

func NewReviewService(b domain.BookRepository, r domain.ReviewRepository) *ReviewService {
	return &ReviewService{
		bookRepo:   b,
		reviewRepo: r,
	}
}

func (s *ReviewService) Create(ctx context.Context, userID, bookID uuid.UUID, input domain.CreateReviewInput) (*domain.Review, error) {
	// Валидация
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("ReviewService.Create: validate input: %w", err)
	}

	// Проверяем наличие книги
	if _, err := s.bookRepo.GetByID(ctx, bookID); err != nil {
		return nil, fmt.Errorf("ReviewService.Create: fetch book: %w", err)
	}

	// Проверяем наличие
	hasReviewed, err := s.reviewRepo.UserHasReviewedBook(ctx, userID, bookID)
	if err != nil {
		return nil, fmt.Errorf("ReviewService.Create: check review: %w", err)
	}
	if hasReviewed {
		return nil, domain.ErrAlreadyReviewed
	}

	review := &domain.Review{
		BookID:  bookID,
		UserID:  userID,
		Rating:  input.Rating,
		Title:   input.Title,
		Content: input.Content,
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("ReviewService.Create: create review: %w", err)
	}

	return review, nil
}

func (s *ReviewService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	// Валидация
	if id == uuid.Nil {
		return nil, ErrEmptyReviewID
	}

	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ReviewService.GetByID: fetch review: %w", err)
	}

	return review, nil
}

func (s *ReviewService) ListByBook(ctx context.Context, bookID uuid.UUID, page, limit int) ([]domain.Review, int, error) {
	// Валидация
	if bookID == uuid.Nil {
		return nil, 0, ErrEmptyBookID
	}

	// defaults
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = DefaultReviewCountLimit
	}

	reviews, rCount, err := s.reviewRepo.ListByBookID(ctx, bookID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("ReviewService.ListByBookID: %w", err)
	}

	return reviews, rCount, nil
}

func (s *ReviewService) Update(ctx context.Context, userID, reviewID uuid.UUID, input *domain.UpdateReviewInput) (*domain.Review, error) {
	// Валидация id
	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}
	if reviewID == uuid.Nil {
		return nil, ErrEmptyReviewID
	}

	// Валидация input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("ReviewService.Update: validate input: %w", err)
	}

	// Получение модели
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, fmt.Errorf("ReviewService.Update: fetch review: %w", err)
	}

	// Проверка владельца
	if review.UserID != userID {
		return nil, domain.ErrNotReviewOwner
	}

	if input.Content != nil {
		review.Content = *input.Content
	}
	if input.Rating != nil {
		review.Rating = *input.Rating
	}
	if input.Title != nil {
		review.Title = input.Title
	}

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, fmt.Errorf("ReviewService.Update: update review: %w", err)
	}

	return review, nil
}

func (s *ReviewService) Delete(ctx context.Context, userID, reviewID uuid.UUID) error {
	// Валидация id
	if userID == uuid.Nil {
		return ErrEmptyUserID
	}
	if reviewID == uuid.Nil {
		return ErrEmptyReviewID
	}

	// Получение модели
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("ReviewService.Delete: fetch review: %w", err)
	}

	// Проверка владельца
	if review.UserID != userID {
		return domain.ErrNotReviewOwner
	}

	if err := s.reviewRepo.Delete(ctx, reviewID); err != nil {
		return fmt.Errorf("ReviewService.Delete: delete review: %w", err)
	}

	return nil
}
