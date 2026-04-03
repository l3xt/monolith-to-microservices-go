package model

import (
	"bookshelf/internal/domain"
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
		ID: u.ID,
		Username: u.Username,
		Email: u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
