package repository

import (
	"bookshelf/internal/database"
	"bookshelf/internal/domain"
	"bookshelf/internal/repository/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ReviewRepository struct {
	db *database.PostgresDB
}

func NewReviewRepository(db *database.PostgresDB) *ReviewRepository {
	if db == nil {
		panic("db is nil")
	}
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	const query = `
		INSERT INTO reviews (id, book_id, user_id, rating, title, content)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := r.db.Pool.Exec(
		ctx,
		query,
		uuid.New(),
		review.BookID,
		review.UserID,
		review.Rating,
		review.Title,
		review.Content,
	); err != nil {
		return fmt.Errorf("ReviewRepository.Create: %w", err)
	}

	return nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	const query = `
		SELECT id, book_id , user_id, rating, title, content, created_by, created_at, updated_at 
		FROM books WHERE id = $1
	`

	rows, err := r.db.Pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("ReviewRepository.GetByID: %w", err)
	}

	reviewDB, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.ReviewDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReviewNotFound
		}
		return nil, fmt.Errorf("ReviewRepository.GetByID: %w", err)
	}
	return reviewDB.ToDomain(), nil
}

func (r *ReviewRepository) ListByBookID(ctx context.Context, bookID uuid.UUID, page, limit int) ([]domain.Review, int, error) {
	const reviewsQuery = `
		SELECT id, book_id , user_id, rating, title, content, created_at, updated_at
		FROM reviews
		WHERE book_id = $1
		LIMIT $2 OFFSET $3
	`
	const countQuery = `SELECT COUNT(*) FROM reviews WHERE book_id = $1`

	offset := (page - 1) * limit

	rows, err := r.db.Pool.Query(ctx, reviewsQuery, bookID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ReviewRepository.ListByBookID: %w", err)
	}

	reviews, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Review, error) {
		var rDB model.ReviewDB
		if err := row.Scan(
			&rDB.ID,
			&rDB.BookID,
			&rDB.UserID,
			&rDB.Rating,
			&rDB.Title,
			&rDB.Content,
			&rDB.CreatedAt,
			&rDB.UpdatedAt,
		); err != nil {
			return domain.Review{}, err
		}

		return *rDB.ToDomain(), nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("ReviewRepository.ListByBookID: %w", err)
	}

	var reviewsCount int
	err = r.db.Pool.QueryRow(ctx, countQuery, bookID).Scan(&reviewsCount)
	if err != nil {
		return nil, 0, fmt.Errorf("ReviewRepository.ListByBookID: %w", err)
	}

	return reviews, reviewsCount, nil
}

func (r *ReviewRepository) Update(ctx context.Context, review *domain.Review) error {
	const query = `
		UPDATE reviews
		SET rating = $1, title = $2, content = $3
		WHERE id = $4
	`
	res, err := r.db.Pool.Exec(ctx, query, review.Rating, review.Title, review.Content, review.ID)
	if err != nil {
		return fmt.Errorf("ReviewRepository.Update: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrReviewNotFound
	}

	return nil
}

func (r *ReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM reviews WHERE id = $1`

	res, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ReviewRepository.Delete: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrReviewNotFound
	}

	return nil
}

func (r *ReviewRepository) UserHasReviewedBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM reviews WHERE user_id = $1 AND book_id = $2)`

	var isExists bool
	err := r.db.Pool.QueryRow(ctx, query, userID, bookID).Scan(&isExists)
	if err != nil {
		return false, fmt.Errorf("ReviewRepository.UserHasReviewedBook: %w", err)
	}

	return isExists, nil
}
