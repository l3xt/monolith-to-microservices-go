package repository

import (
	"bookshelf/users-service/internal/domain"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type UserDB struct {
	ID           uuid.UUID `db:"id"`
	Username     string    `db:"username"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (u *UserDB) ToDomain() *domain.User {
	return &domain.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// Конвертирует ссылочный тип в sql.Null
func ToNull[T any](v *T) sql.Null[T] {
	if v == nil {
		return sql.Null[T]{}
	}
	return sql.Null[T]{V: *v, Valid: true}
}

// Конвертирует sql.Null в ссылочный тип
func FromNull[T any](v sql.Null[T]) *T {
	if !v.Valid {
		return nil
	}
	return &v.V
}
