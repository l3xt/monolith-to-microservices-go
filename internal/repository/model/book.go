package model

import (
	"bookshelf/internal/domain"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type BookDB struct {
	ID            uuid.UUID        `db:"id"`
	Title         string           `db:"title"`
	Author        string           `db:"author"`
	Description   sql.Null[string] `db:"description"`
	ISBN          sql.Null[string] `db:"isbn"`
	PublishedYear sql.Null[int32]  `db:"published_year"`
	AverageRating float64          `db:"average_rating"` // считается
	ReviewsCount  int              `db:"reviews_count"`  // считается
	CreatedBy     uuid.UUID        `db:"created_by"`
	CreatedAt     time.Time        `db:"created_at"`
	UpdatedAt     time.Time        `db:"updated_at"`
}

func (b *BookDB) ToDomain() *domain.Book {
	return &domain.Book{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		Description:   FromNull(b.Description),
		ISBN:          FromNull(b.ISBN),
		PublishedYear: FromNull(b.PublishedYear),
		AverageRating: b.AverageRating,
		ReviewsCount:  b.ReviewsCount,
		CreatedBy:     b.CreatedBy,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}
