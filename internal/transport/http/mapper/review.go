package mapper

import (
	"bookshelf/internal/domain"
	"bookshelf/internal/transport/http/dto"
)

func ToReviewResponse(r *domain.Review, user *dto.UserSummary) *dto.ReviewResponse {
	return &dto.ReviewResponse{
		ID:        r.ID,
		BookID:    r.BookID,
		User:      user,
		Rating:    r.Rating,
		Title:     r.Title,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
