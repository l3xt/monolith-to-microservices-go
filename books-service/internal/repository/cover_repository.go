package repository

import (
	"bookshelf/books-service/internal/database"
	"bookshelf/books-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CoverRepository struct {
	db *database.PostgresDB
}

func NewCoverRepository(db *database.PostgresDB) *CoverRepository {
	if db == nil {
		panic("NewCoverRepository: db is nil")
	}
	return &CoverRepository{db: db}
}

func (r *CoverRepository) Create(ctx context.Context, cover *domain.Cover) error {
	const query = `
		INSERT INTO covers (book_id, status, original_path, cover_path, thumb_path, error, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	coverDB := NewCoverDB(cover)

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		coverDB.BookID,
		coverDB.Status,
		coverDB.OriginalPath,
		coverDB.CoverPath,
		coverDB.ThumbPath,
		coverDB.Error,
		coverDB.CompletedAt,
	).Scan(&cover.ID, &cover.CreatedAt, &cover.UpdatedAt)

	if err != nil {
		return fmt.Errorf("CoverRepository.Create: %w", err)
	}

	return nil
}

func (r *CoverRepository) GetByBookID(ctx context.Context, bookID uuid.UUID) (*domain.Cover, error) {
	const query = `
		SELECT id, book_id, status, original_path, cover_path, thumb_path, error, created_at, updated_at, completed_at
		FROM covers
		WHERE book_id = $1
		LIMIT 1
	`

	rows, err := r.db.Pool.Query(ctx, query, bookID)
	if err != nil {
		return nil, fmt.Errorf("CoverRepository.GetByBookID: %w", err)
	}

	coverDB, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[CoverDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCoverNotFound
		}
		return nil, fmt.Errorf("CoverRepository.GetByBookID: %w", err)
	}

	return coverDB.ToDomain(), nil
}

func (r *CoverRepository) UpdateStatus(ctx context.Context, id string, status domain.CoverStatus, coverPath, thumbPath, errorMsg string) error {
	const query = `
		UPDATE covers 
		SET status = $1, cover_path = $2, thumb_path = $3, error = $4, completed_at = $5
		WHERE id = $6
	`

	var completedAt *time.Time
	if status == domain.CoverStatusReady || status == domain.CoverStatusFailed {
		now := time.Now()
		completedAt = &now
	}

	_, err := r.db.Pool.Exec(ctx, query, status, coverPath, thumbPath, errorMsg, completedAt, id)
	if err != nil {
		return fmt.Errorf("CoverRepository.UpdateStatus: %w", err)
	}

	return nil
}

func (r *CoverRepository) DeleteByBookID(ctx context.Context, bookID uuid.UUID) error {
	const query = `DELETE FROM covers WHERE book_id = $1`

	_, err := r.db.Pool.Exec(ctx, query, bookID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("CoverRepository.DeleteByBookID: %w", domain.ErrCoverNotFound)
		}
		return fmt.Errorf("CoverRepository.DeleteByBookID: %w", err)
	}

	return nil
}
