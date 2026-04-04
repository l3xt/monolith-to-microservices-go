package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReviewResponse struct {
	ID        uuid.UUID `json:"id"`
	BookID    uuid.UUID `json:"book_id"`
	UserID    uuid.UUID `json:"user_id"`
	Rating    int       `json:"rating"`
	Title     *string   `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReviewListResponse struct {
	Data       []ReviewResponse `json:"data"`
	Pagination *Pagination      `json:"pagination"`
}

type CreateReviewRequest struct {
	Title   *string `json:"title"`
	Content string  `json:"content"`
	Rating  int     `json:"rating"`
}

type UpdateReviewRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	Rating  *int    `json:"rating"`
}
