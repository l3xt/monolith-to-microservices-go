package repository

import (
	"bookshelf/books-service/internal/database"
	"bookshelf/books-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var allowedSortFields = map[string]string{
	"title":          "title",
	"author":         "author",
	"published_year": "published_year",
	"rating":         "average_rating",
	"updated_at":     "updated_at",
	"created_at":     "created_at",
}

type BookRepository struct {
	db *database.PostgresDB
}

func NewBookRepository(db *database.PostgresDB) *BookRepository {
	if db == nil {
		panic("db is nil")
	}
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, book *domain.Book) error {
	const query = `
		INSERT INTO books (title, author, description, isbn, published_year, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		book.Title,
		book.Author,
		book.Description,
		book.ISBN,
		book.PublishedYear,
		book.UserID,
	).Scan(&book.ID, &book.CreatedAt, &book.UpdatedAt)

	if err != nil {
		return fmt.Errorf("BookRepository.Create: %w", err)
	}

	return nil
}

func (r *BookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	const query = `
		SELECT 
			b.id, 
			b.title, 
			b.author, 
			b.description, 
			b.isbn, 
			b.published_year, 
			b.user_id, 
			b.created_at, 
			b.updated_at,
			COALESCE(ROUND(AVG(r.rating), 2), 0) AS average_rating,
			COUNT(r) AS reviews_count
		FROM books b 
		LEFT JOIN reviews r ON r.book_id = b.id
		WHERE b.id = $1
		GROUP BY b.id
	`

	rows, err := r.db.Pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("BookRepository.GetByID: %w", err)
	}

	bookDB, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[BookDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookNotFound
		}
		return nil, fmt.Errorf("BookRepository.GetByID: %w", err)
	}

	return bookDB.ToDomain(), nil
}

func (r *BookRepository) List(ctx context.Context, filter *domain.BookFilter) ([]domain.Book, int, error) {
	column := pgx.Identifier{allowedSortFields[filter.Sort]}.Sanitize()
	direction := "ASC"
	if strings.ToUpper(filter.Order) == "DESC" {
		direction = "DESC"
	}

	searchQuery := `
		SELECT 
			b.id, 
			b.title, 
			b.author, 
			b.description, 
			b.isbn, 
			b.published_year, 
			b.user_id, 
			b.created_at, 
			b.updated_at,
			COALESCE(ROUND(AVG(r.rating), 2), 0) AS average_rating,
			COUNT(r.id) AS reviews_count
		FROM books b
		LEFT JOIN reviews r ON r.book_id = b.id
		WHERE $1 = '' OR b.title ILIKE $1 OR b.author ILIKE $1
		GROUP BY b.id
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`
	searchQuery = fmt.Sprintf(searchQuery, column, direction)

	countQuery := `SELECT COUNT(*) FROM books WHERE $1 = '' OR title ILIKE $1 OR author ILIKE $1`

	offset := (filter.Page - 1) * filter.Limit

	// Отправляем оба запроса за один round trip
	batch := &pgx.Batch{}
	batch.Queue(countQuery, filter.Search)
	batch.Queue(searchQuery, filter.Search, filter.Limit, offset)

	br := r.db.Pool.SendBatch(ctx, batch)
	defer br.Close()

	var booksCount int
	if err := br.QueryRow().Scan(&booksCount); err != nil {
		return nil, 0, fmt.Errorf("BookRepository.List: %w", err)
	}

	rows, err := br.Query()
	if err != nil {
		return nil, 0, fmt.Errorf("BookRepository.List: %w", err)
	}
	defer rows.Close()

	books, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Book, error) {
		var bDB BookDB
		if err := row.Scan(
			&bDB.ID, &bDB.Title, &bDB.Author, &bDB.Description,
			&bDB.ISBN, &bDB.PublishedYear, &bDB.UserID,
			&bDB.CreatedAt, &bDB.UpdatedAt,
			&bDB.AverageRating, &bDB.ReviewsCount,
		); err != nil {
			return domain.Book{}, err
		}
		return *bDB.ToDomain(), nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("BookRepository.List: %w", err)
	}

	return books, booksCount, nil
}

func (r *BookRepository) ListByUser(ctx context.Context, userID uuid.UUID, filter *domain.BookFilter) ([]domain.Book, int, error) {
column := pgx.Identifier{allowedSortFields[filter.Sort]}.Sanitize()
	direction := "ASC"
	if strings.ToUpper(filter.Order) == "DESC" {
		direction = "DESC"
	}

	searchQuery := `
		SELECT 
			b.id, 
			b.title, 
			b.author, 
			b.description, 
			b.isbn, 
			b.published_year, 
			b.user_id, 
			b.created_at, 
			b.updated_at,
			COALESCE(ROUND(AVG(r.rating), 2), 0) AS average_rating,
			COUNT(r.id) AS reviews_count
		FROM books b
		LEFT JOIN reviews r ON r.book_id = b.id
		WHERE b.user_id = $1 AND $2 = '' OR b.title ILIKE $2 OR b.author ILIKE $2
		GROUP BY b.id
		ORDER BY %s %s
		LIMIT $3 OFFSET $4
	`
	searchQuery = fmt.Sprintf(searchQuery, column, direction)

	countQuery := `SELECT COUNT(*) FROM books WHERE $1 = '' OR title ILIKE $1 OR author ILIKE $1`

	offset := (filter.Page - 1) * filter.Limit

	// Отправляем оба запроса за один round trip
	batch := &pgx.Batch{}
	batch.Queue(countQuery, filter.Search)
	batch.Queue(searchQuery, filter.Search, filter.Limit, offset)

	br := r.db.Pool.SendBatch(ctx, batch)
	defer br.Close()

	var booksCount int
	if err := br.QueryRow().Scan(&booksCount); err != nil {
		return nil, 0, fmt.Errorf("BookRepository.List: %w", err)
	}

	rows, err := br.Query()
	if err != nil {
		return nil, 0, fmt.Errorf("BookRepository.List: %w", err)
	}
	defer rows.Close()

	books, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Book, error) {
		var bDB BookDB
		if err := row.Scan(
			&bDB.ID, &bDB.Title, &bDB.Author, &bDB.Description,
			&bDB.ISBN, &bDB.PublishedYear, &bDB.UserID,
			&bDB.CreatedAt, &bDB.UpdatedAt,
			&bDB.AverageRating, &bDB.ReviewsCount,
		); err != nil {
			return domain.Book{}, err
		}
		return *bDB.ToDomain(), nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("BookRepository.List: %w", err)
	}

	return books, booksCount, nil
}

func (r *BookRepository) Update(ctx context.Context, book *domain.Book) error {
	const query = `
		UPDATE books
		SET title = $1, author = $2, description = $3, isbn = $4, published_year = $5
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		book.Title,
		book.Author,
		book.Description,
		book.ISBN,
		book.PublishedYear,
		book.ID,
	).Scan(&book.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrBookNotFound
		}
		return fmt.Errorf("BookRepository.Update: %w", err)
	}

	return nil
}

func (r *BookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM books WHERE id = $1`

	res, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("BookRepository.Delete: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrBookNotFound
	}

	return nil
}
