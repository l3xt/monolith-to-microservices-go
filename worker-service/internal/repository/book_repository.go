package repository

import (
	"bookshelf/worker-service/internal/database"
	"bookshelf/worker-service/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BookRepository struct {
	db *database.PostgresDB
}

func NewBookRepository(db *database.PostgresDB) *BookRepository {
	if db == nil {
		panic("NewBookRepository: db is nil")
	}
	return &BookRepository{db: db}
}

func (r *BookRepository) UpdateCover(ctx context.Context, bookID uuid.UUID, coverURL, thumbURL *string, status domain.CoverStatus) error {
	const query = `
		UPDATE books
		SET cover_status = $1, cover_url = $2, thumbnail_url = $3
		WHERE id = $4
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		status,
		coverURL,
		thumbURL,
		bookID,
	)

	if err != nil {
		if r.db.IsRetryable(err) {
			err = fmt.Errorf("%w: %w", domain.ErrServiceNotResponding, err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			err = domain.ErrBookNotFound
		}
		return fmt.Errorf("BookRepository.UpdateCover: %w", err)
	}

	return nil
}
