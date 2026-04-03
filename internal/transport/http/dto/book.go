package dto

import (
	"time"

	"github.com/google/uuid"
)

type BookResponse struct {
	ID            uuid.UUID   `db:"id" json:"id"`
	Title         string      `db:"title" json:"title"`
	Author        string      `db:"author" json:"author"`
	Description   *string     `db:"description" json:"description"`
	ISBN          *string     `db:"isbn" json:"isbn"`
	PublisherYear *int32      `db:"published_year" json:"published_year"`
	AverageRating float64     `db:"-" json:"average_rating"`
	ReviewsCount  int         `db:"-" json:"reviews_count"`
	CreatedBy     *UserSummary `db:"created_by" json:"created_by"`
	CreatedAt     time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at" json:"updated_at"`
}

type BookListResponse struct {
	Data       []BookResponse `json:"data"`
	Pagination *Pagination    `json:"pagination"`
}

type CreateBookRequest struct {
	Title         string  `db:"title" json:"title"`
	Author        string  `db:"author" json:"author"`
	Description   *string `db:"description" json:"description"`
	ISBN          *string `db:"isbn" json:"isbn"`
	PublishedYear *int32  `db:"published_year" json:"published_year"`
}

type UpdateBookRequest struct {
	Title         *string `db:"title" json:"title"`
	Author        *string `db:"author" json:"author"`
	Description   *string `db:"description" json:"description"`
	ISBN          *string `db:"isbn" json:"isbn"`
	PublishedYear *int32  `db:"published_year" json:"published_year"`
}
