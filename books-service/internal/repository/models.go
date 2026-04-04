package repository

import (
	"bookshelf/books-service/internal/domain"
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
	UserID        uuid.UUID        `db:"user_id"`
	CreatedAt     time.Time        `db:"created_at"`
	UpdatedAt     time.Time        `db:"updated_at"`
}

type ReviewDB struct {
	ID        uuid.UUID        `db:"id"`
	BookID    uuid.UUID        `db:"book_id"`
	UserID    uuid.UUID        `db:"user_id"`
	Rating    int              `db:"rating"`
	Title     sql.Null[string] `db:"title"`
	Content   string           `db:"content"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
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
		UserID:        b.UserID,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

func (r *ReviewDB) ToDomain() *domain.Review {
	return &domain.Review{
		ID:        r.ID,
		BookID:    r.BookID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Title:     FromNull(r.Title),
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
