package mapper

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/transport/http/dto"
)

func ToReviewResponse(r *domain.Review) *dto.ReviewResponse {
	return &dto.ReviewResponse{
		ID:        r.ID,
		BookID:    r.BookID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Title:     r.Title,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
