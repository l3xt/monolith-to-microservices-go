package repository

import (
	"bookshelf/worker-service/internal/database"
	"bookshelf/worker-service/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (r *CoverRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CoverStatus, coverPath, thumbPath, errorMsg *string) error {
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
