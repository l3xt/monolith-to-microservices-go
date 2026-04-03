package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReviewResponse struct {
	ID        uuid.UUID   `db:"id" json:"id"`
	BookID    uuid.UUID   `db:"book_id" json:"book_id"`
	User      *UserSummary `db:"user" json:"user"`
	Rating    int         `db:"rating" json:"rating"`
	Title     *string     `db:"title" json:"title"`
	Content   string      `db:"content" json:"content"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
}

type ReviewListResponse struct {
	Data       []ReviewResponse `json:"data"`
	Pagination *Pagination       `json:"pagination"`
}

type CreateReviewRequest struct {
	Title   *string `db:"title" json:"title"`
	Content string  `db:"content" json:"content"`
	Rating  int     `db:"rating" json:"rating"`
}

type UpdateReviewRequest struct {
	Title   *string `db:"title" json:"title"`
	Content *string `db:"content" json:"content"`
	Rating  *int    `db:"rating" json:"rating"`
}