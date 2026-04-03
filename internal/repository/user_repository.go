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
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db *database.PostgresDB
}

func NewUserRepository(db *database.PostgresDB) *UserRepository {
	if db == nil {
		panic("db is nil")
	}
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (id, username, email, password_hash) 
		VALUES ($1, $2, $3, $4)
	`

	if _, err := r.db.Pool.Exec(
		ctx,
		query,
		uuid.New(),
		user.Username,
		user.Email,
		user.PasswordHash,
	); err != nil {
		var pgErr *pgconn.PgError
		// SQLSTATE 23505 - unique_violation в PostgreSQL
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return domain.ErrUserExists
			case "users_username_key":
				return domain.ErrUsernameExists
			}
		}
		return fmt.Errorf("UserRepository.Create: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT id, username, email, password_hash, created_at, updated_at 
		FROM users WHERE id = $1
	`

	rows, err := r.db.Pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetByID: %w", err)
	}

	userDB, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.UserDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByID: %w", err)
	}
	return userDB.ToDomain(), nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}

	const query = `
		SELECT id, username, email, password_hash, created_at, updated_at 
		FROM users WHERE id = ANY($1)
	`

	rows, err := r.db.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetByIDs: %w", err)
	}

	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.User, error) {
		var uDB model.UserDB
		if err := row.Scan(
			&uDB.ID,
			&uDB.Username,
			&uDB.Email,
			&uDB.PasswordHash,
			&uDB.CreatedAt,
			&uDB.UpdatedAt,
		); err != nil {
			return domain.User{}, err
		}
		return *uDB.ToDomain(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetByIDs: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, username, email, password_hash, created_at, updated_at 
		FROM users WHERE email = $1
	`

	rows, err := r.db.Pool.Query(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetByEmail: %w", err)
	}

	userDB, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.UserDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByEmail: %w", err)
	}

	return userDB.ToDomain(), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	const query = `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users WHERE username = $1
	`

	rows, err := r.db.Pool.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetByUsername: %w", err)
	}

	userDB, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.UserDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByUsername: %w", err)
	}

	return userDB.ToDomain(), nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	const query = `
		UPDATE users 
		SET username = $1, email = $2, password_hash = $3
		WHERE id = $4
	`

	res, err := r.db.Pool.Exec(ctx, query, user.Username, user.Email, user.PasswordHash, user.ID)
	if err != nil {
		return fmt.Errorf("UserRepository.Update: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`

	var isExists bool
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(&isExists)
	if err != nil {
		return false, fmt.Errorf("UserRepository.EmailExists: %w", err)
	}

	return isExists, nil
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)`

	var isExists bool
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(&isExists)
	if err != nil {
		return false, fmt.Errorf("UserRepository.UsernameExists: %w", err)
	}

	return isExists, nil
}
