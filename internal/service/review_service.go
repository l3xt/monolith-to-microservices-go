package service

import (
	"bookshelf/internal/domain"
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MinRating          = 1
	MaxRating          = 5
	ContentMinLen      = 10
	DefaultReviewLimit = 20
)

var (
	ErrReviewNotFound        = errors.New("review not found")
	ErrNotReviewOwner        = errors.New("user is not review owner")
	ErrAlreadyReviewed       = errors.New("user has already left a review about this book")
	ErrInvalidRating         = errors.New("rating value is not in the range from 1 to 5")
	ErrReviewContentTooShort = errors.New("content length must be more than 10 characters")
)

type ReviewService struct {
	reviewRepo domain.ReviewRepository
	bookRepo   domain.BookRepository
	userRepo   domain.UserRepository
}

func NewReviewService(u domain.UserRepository, b domain.BookRepository, r domain.ReviewRepository) *ReviewService {
	return &ReviewService{
		userRepo:   u,
		bookRepo:   b,
		reviewRepo: r,
	}
}

func (s *ReviewService) Create(ctx context.Context, userID, bookID uuid.UUID, in domain.CreateReviewInput) (*domain.Review, *domain.User, error) {
	// Валидация
	if in.Rating < MinRating || in.Rating > MaxRating {
		return nil, nil, ErrInvalidRating
	}
	if utf8.RuneCountInString(in.Content) < ContentMinLen {
		return nil, nil, ErrReviewContentTooShort
	}

	// Проверяем наличие книги
	if _, err := s.bookRepo.GetByID(ctx, bookID); err != nil {
		return nil, nil, fmt.Errorf("ReviewService.Create: %w", err)
	}

	// Проверяем наличие
	hasReviewed, err := s.reviewRepo.UserHasReviewedBook(ctx, userID, bookID)
	if err != nil {
		return nil, nil, fmt.Errorf("ReviewService.Create: %w", err)
	}
	if hasReviewed {
		return nil, nil, ErrAlreadyReviewed
	}

	review := domain.Review{
		BookID:  bookID,
		UserID:  userID,
		Rating:  in.Rating,
		Title:   in.Title,
		Content: in.Content,
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("ReviewService.Create: %w", err)
	}

	if err := s.reviewRepo.Create(ctx, &review); err != nil {
		return nil, nil, fmt.Errorf("ReviewService.Create: %w", err)
	}

	return &review, user, nil
}

func (s *ReviewService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, *domain.User, error) {
	if id == uuid.Nil {
		return nil, nil, ErrEmptyID
	}

	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("ReviewService.GetByID: %w", err)
	}

	// Получаем пользователя для summary
	user, err := s.userRepo.GetByID(ctx, review.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("ReviewService.GetByID: %w", err)
	}

	return review, user, nil
}

func (s *ReviewService) ListByBookID(ctx context.Context, bookID uuid.UUID, page, limit int) ([]domain.Review, map[uuid.UUID]*domain.User, int, error) {
	// Валидация
	if bookID == uuid.Nil {
		return nil, nil, 0, ErrEmptyID
	}

	// defaults
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = DefaultReviewLimit
	}

	reviews, rCount, err := s.reviewRepo.ListByBookID(ctx, bookID, page, limit)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("ReviewService.ListByBookID: %w", err)
	}
	if len(reviews) == 0 {
		return []domain.Review{}, nil, 0, nil
	}

	// Собираем уникальных авторов
	creatorIDs := make([]uuid.UUID, 0, rCount)
	seen := make(map[uuid.UUID]struct{})
	for _, r := range reviews {
		if _, ok := seen[r.UserID]; !ok {
			creatorIDs = append(creatorIDs, r.UserID)
			seen[r.UserID] = struct{}{}
		}
	}

	// Получаем данные об этих авторах
	users, err := s.userRepo.GetByIDs(ctx, creatorIDs)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("ReviewService.ListByBookID: %w", err)
	}

	// Создаем мапу для O(1) доступа
	usersMap := make(map[uuid.UUID]*domain.User, len(users))
	for _, u := range users {
		usersMap[u.ID] = &u
	}

	return reviews, usersMap, rCount, nil
}

func (s *ReviewService) Update(ctx context.Context, userID, reviewID uuid.UUID, in *domain.UpdateReviewInput) (*domain.Review, *domain.User, error) {
	// Валидация id
	if userID == uuid.Nil || reviewID == uuid.Nil {
		return nil, nil, ErrEmptyID
	}

	// Валидация input
	if in == nil {
		return nil, nil, ErrInvalidInputValue
	}
	if in.Content == nil && in.Rating == nil && in.Title == nil {
		return nil, nil, ErrInvalidInputValue
	}
	if in.Content != nil && utf8.RuneCountInString(*in.Content) < ContentMinLen {
		return nil, nil, ErrReviewContentTooShort
	}
	if in.Rating != nil && (*in.Rating < MinRating || *in.Rating > MaxRating) {
		return nil, nil, ErrInvalidRating
	}

	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, nil, fmt.Errorf("ReviewService.Update: %w", err)
	}

	// Проверка владельца
	if review.UserID != userID {
		return nil, nil, ErrNotReviewOwner
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("ReviewService.Update: %w", err)
	}

	if in.Content != nil {
		review.Content = *in.Content
	}
	if in.Rating != nil {
		review.Rating = *in.Rating
	}
	if in.Title != nil {
		review.Title = in.Title
	}

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("ReviewService.Update: %w", err)
	}

	return review, user, nil
}

func (s *ReviewService) Delete(ctx context.Context, userID, reviewID uuid.UUID) error {
	// Валидация id
	if userID == uuid.Nil || reviewID == uuid.Nil {
		return ErrEmptyID
	}

	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("ReviewService.Delete: %w", err)
	}

	// Проверка владельца
	if userID != review.UserID {
		return ErrNotReviewOwner
	}

	if err := s.reviewRepo.Delete(ctx, reviewID); err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return err
		}
		return fmt.Errorf("ReviewService.Delete: %w", err)
	}
	return nil
}
