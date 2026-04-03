package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrReviewNotFound = errors.New("review not found")

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

type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*Review, error)
	ListByBookID(ctx context.Context, bookID uuid.UUID, page, limit int) ([]Review, int, error)
	Update(ctx context.Context, review *Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	UserHasReviewedBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error)
}
