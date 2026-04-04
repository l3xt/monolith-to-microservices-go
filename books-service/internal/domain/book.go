package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultBookLimit = 100
	DefaultBookSort  = "title"
)

var (
	ErrInvalidInputValue      = errors.New("invalid input value")
	ErrInvalidBookFilterValue = errors.New("invalid book filter value")
	ErrInvalidSortValue       = errors.New("invalid sort value in filter object")
	ErrInvalidOrderValue      = errors.New("invalid order value in filter object")
	ErrBookNotFound           = errors.New("book not found")
	ErrBookTitleEmpty         = errors.New("empty book title")
	ErrBookAuthorEmpty        = errors.New("empty book author")
	ErrNotBookOwner           = errors.New("not book owner")
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
	UserID        uuid.UUID
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
	if f.Limit <= 0 {
		return ErrInvalidBookFilterValue
	}
	if f.Page <= 0 {
		return ErrInvalidBookFilterValue
	}

	return nil
}

func (f *BookFilter) GetOffset() int {
	return (f.Page - 1) * f.Limit
}

func (f *BookFilter) SetDefaults() {
	if f.Limit <= 0 {
		f.Limit = DefaultBookLimit
	}
	if f.Sort == "" {
		f.Sort = DefaultBookSort
	}
	if f.Page <= 0 {
		f.Page = 1
	}
}

func (input *CreateBookInput) Validate() error {
	if input == nil {
		return ErrInvalidInputValue
	}
	if input.Title == "" {
		return ErrBookTitleEmpty
	}
	if input.Author == "" {
		return ErrBookAuthorEmpty
	}
	return nil
}

func (input *UpdateBookInput) Validate() error {
	if input == nil {
		return ErrInvalidInputValue
	}
	if input.Title == nil && input.Author == nil && input.Description == nil && input.ISBN == nil && input.PublishedYear == nil {
		return ErrInvalidInputValue
	}
	if input.Title != nil && *input.Title == "" {
		return ErrBookTitleEmpty
	}
	if input.Author != nil && *input.Author == "" {
		return ErrBookAuthorEmpty
	}
	if input.PublishedYear != nil && *input.PublishedYear <= 0 {
		return ErrInvalidInputValue
	}
	if input.ISBN != nil && *input.ISBN == "" {
		return ErrInvalidInputValue
	}
	if input.Description != nil && *input.Description == "" {
		return ErrInvalidInputValue
	}

	return nil
}
