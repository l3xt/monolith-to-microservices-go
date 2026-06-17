package dto

import (
	"bookshelf/books-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type BookResponse struct {
	ID            uuid.UUID          `json:"id"`
	Title         string             `json:"title"`
	Author        string             `json:"author"`
	Description   *string            `json:"description"`
	ISBN          *string            `json:"isbn"`
	PublisherYear *int32             `json:"published_year"`
	AverageRating float64            `json:"average_rating"`
	ReviewsCount  int                `json:"reviews_count"`
	UserID        uuid.UUID          `json:"created_by"`
	CoverStatus   domain.CoverStatus `json:"cover_status"`
	CoverURL      *string            `json:"cover_url,omitempty"`
	ThumbURL      *string            `json:"thumb_url,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}
type BookListResponse struct {
	Data       []BookResponse `json:"data"`
	Pagination *Pagination    `json:"pagination"`
}

type CreateBookRequest struct {
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Description   *string `json:"description"`
	ISBN          *string `json:"isbn"`
	PublishedYear *int32  `json:"published_year"`
}

type UpdateBookRequest struct {
	Title         *string `json:"title"`
	Author        *string `json:"author"`
	Description   *string `json:"description"`
	ISBN          *string `json:"isbn"`
	PublishedYear *int32  `json:"published_year"`
}
