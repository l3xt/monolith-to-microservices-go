package mapper

import (
	"bookshelf/internal/domain"
	"bookshelf/internal/transport/http/dto"
)

func ToBookResponse(b *domain.Book, creator *dto.UserSummary) *dto.BookResponse {
	return &dto.BookResponse{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		Description:   b.Description,
		ISBN:          b.ISBN,
		PublisherYear: b.PublishedYear,
		AverageRating: b.AverageRating,
		ReviewsCount:  b.ReviewsCount,
		CreatedBy:     creator,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}
