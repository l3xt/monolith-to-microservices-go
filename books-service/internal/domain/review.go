package domain

import (
	"context"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	ReviewMinRating     = 1
	ReviewMaxRating     = 5
	ReviewContentMinLen = 10
)

type Review struct {
	ID        uuid.UUID
	BookID    uuid.UUID
	UserID    uuid.UUID
	Rating    int
	Title     *string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateReviewInput struct {
	Title   *string
	Content string
	Rating  int
}

type UpdateReviewInput struct {
	Title   *string
	Content *string
	Rating  *int
}

func (input *CreateReviewInput) Validate() error {
	if input == nil {
		return ErrInvalidInputValue
	}
	if input.Rating < ReviewMinRating || input.Rating > ReviewMaxRating {
		return ErrInvalidRating
	}
	if utf8.RuneCountInString(input.Content) < ReviewContentMinLen {
		return ErrReviewContentTooShort
	}
	return nil
}

func (input *UpdateReviewInput) Validate() error {
	if input == nil {
		return ErrInvalidInputValue
	}
	if input.Title == nil && input.Rating == nil && input.Content == nil {
		return ErrInvalidInputValue
	}
	if input.Rating != nil && (*input.Rating < ReviewMinRating || *input.Rating > ReviewMaxRating) {
		return ErrInvalidRating
	}
	if input.Content != nil && utf8.RuneCountInString(*input.Content) < ReviewContentMinLen {
		return ErrReviewContentTooShort
	}
	return nil
}

type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*Review, error)
	ListByBookID(ctx context.Context, bookID uuid.UUID, page, limit int) ([]Review, int, error)
	Update(ctx context.Context, review *Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	UserHasReviewedBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error)
}
