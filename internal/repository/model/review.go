package model

import (
	"bookshelf/internal/domain"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

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
