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
	CoverStatus   sql.Null[string] `db:"cover_status"`
	CoverURL      sql.Null[string] `db:"cover_url"`
	ThumbnailURL  sql.Null[string] `db:"thumbnail_url"`
	UserID        uuid.UUID        `db:"user_id"`
	CreatedAt     time.Time        `db:"created_at"`
	UpdatedAt     time.Time        `db:"updated_at"`
}

// Обложка для книг
type CoverDB struct {
	ID           uuid.UUID           `db:"id"`
	BookID       uuid.UUID           `db:"book_id"`
	Status       string              `db:"status"`
	OriginalPath sql.Null[string]    `db:"original_path"`
	CoverPath    sql.Null[string]    `db:"cover_path"`
	ThumbPath    sql.Null[string]    `db:"thumb_path"`
	Error        sql.Null[string]    `db:"error"`
	CreatedAt    time.Time           `db:"created_at"`
	UpdatedAt    time.Time           `db:"updated_at"`
	CompletedAt  sql.Null[time.Time] `db:"completed_at"`
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

func NewBookDB(b *domain.Book) *BookDB {
	var coverStatus, coverURL, thumbURL *string
	if b.Cover != nil {
		status := string(b.Cover.Status)
		coverStatus = &status
		coverURL = &b.Cover.CoverURL
		thumbURL = &b.Cover.ThumbURL
	}

	return &BookDB{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		Description:   ToNull(b.Description),
		ISBN:          ToNull(b.ISBN),
		PublishedYear: ToNull(b.PublishedYear),
		AverageRating: b.AverageRating,
		ReviewsCount:  b.ReviewsCount,
		CoverStatus:   ToNull(coverStatus),
		CoverURL:      ToNull(coverURL),
		ThumbnailURL:  ToNull(thumbURL),
		UserID:        b.UserID,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
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

func NewCoverDB(c *domain.Cover) *CoverDB {
	return &CoverDB{
		ID:           c.ID,
		BookID:       c.BookID,
		Status:       string(c.Status),
		OriginalPath: ToNull(c.OriginalPath),
		CoverPath:    ToNull(c.CoverPath),
		ThumbPath:    ToNull(c.ThumbPath),
		Error:        ToNull(c.Error),
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		CompletedAt:  ToNull(c.CompletedAt),
	}
}

func (c *CoverDB) ToDomain() *domain.Cover {
	return &domain.Cover{
		ID:           c.ID,
		BookID:       c.BookID,
		Status:       domain.CoverStatus(c.Status),
		OriginalPath: FromNull(c.OriginalPath),
		CoverPath:    FromNull(c.CoverPath),
		ThumbPath:    FromNull(c.ThumbPath),
		Error:        nil,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		CompletedAt:  FromNull(c.CompletedAt),
	}
}

func NewReviewDB(r *domain.Review) *ReviewDB {
	return &ReviewDB{
		ID:        r.ID,
		BookID:    r.BookID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Title:     ToNull(r.Title),
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
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
