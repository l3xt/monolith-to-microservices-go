package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidBookFilterValue = errors.New("invalid book filter value")
	ErrInvalidSortValue       = errors.New("invalid sort value in filter object")
	ErrInvalidOrderValue      = errors.New("invalid order value in filter object")
	ErrBookNotFound           = errors.New("book not found")
)

// MODELS
// Основная для хранения
type Book struct {
	ID            uuid.UUID
	Title         string
	Author        string
	Description   *string
	ISBN          *string
	PublishedYear *int32
	AverageRating float64
	ReviewsCount  int
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type BookFilter struct {
	Search string
	Sort   string
	Order  string
	Page   int
	Limit  int
}

type CreateBookInput struct {
	Title         string
	Author        string
	Description   *string
	ISBN          *string
	PublishedYear *int32
}

type UpdateBookInput struct {
	Title         *string
	Author        *string
	Description   *string
	ISBN          *string
	PublishedYear *int32
}

type BookRepository interface {
	Create(ctx context.Context, book *Book) error
	GetByID(ctx context.Context, id uuid.UUID) (*Book, error)
	List(ctx context.Context, filter *BookFilter) ([]Book, int, error)
	Update(ctx context.Context, book *Book) error
	Delete(ctx context.Context, id uuid.UUID) error
}

func (f *BookFilter) Validate() error {
	if f == nil {
		return ErrInvalidBookFilterValue
	}
	if f.Order == "" {
		return ErrInvalidOrderValue
	}
	if f.Sort == "" {
		return ErrInvalidSortValue
	}
	return nil
}
