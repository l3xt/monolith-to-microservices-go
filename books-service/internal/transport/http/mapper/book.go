package mapper

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/transport/http/dto"
)

func ToBookResponse(b *domain.Book) *dto.BookResponse {
	return &dto.BookResponse{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		Description:   b.Description,
		ISBN:          b.ISBN,
		PublisherYear: b.PublishedYear,
		AverageRating: b.AverageRating,
		ReviewsCount:  b.ReviewsCount,
		UserID:        b.UserID,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}
