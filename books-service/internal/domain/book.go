package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	DefaultBookLimit = 100
	DefaultBookSort  = "title"
)

// БИЗНЕС МОДЕЛИ
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
	Cover         *Cover // Обложка
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
