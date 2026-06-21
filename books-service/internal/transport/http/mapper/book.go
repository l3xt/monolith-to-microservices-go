package mapper

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/transport/http/dto"
)

func ToBookResponse(b *domain.Book) *dto.BookResponse {
	resp := &dto.BookResponse{
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
		CoverStatus:   domain.CoverStatusNone,
	}

	if b.Cover != nil {
		if b.Cover.Status != "" {
			resp.CoverStatus = b.Cover.Status
		}
		resp.CoverURL = b.Cover.CoverURL
		resp.ThumbURL = b.Cover.ThumbURL
	}

	return resp
}
