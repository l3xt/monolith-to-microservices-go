package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrUserNotFound       = errors.New("user not found")
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type UpdateUserInput struct {
	Username *string
	Email    *string
	Password *string
}

func NewUser(username, email, hash string) *User {
	return &User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
	}
}

func (i *RegisterInput) Validate() error {
	if i.Username == "" {
		return ErrInvalidUsername
	}
	if i.Email == "" {
		return ErrInvalidEmail
	}
	if i.Password == "" {
		return ErrInvalidPassword
	}
	return nil
}

func (i *LoginInput) Validate() error {
	if i.Email == "" {
		return ErrInvalidEmail
	}
	if i.Password == "" {
		return ErrInvalidPassword
	}
	return nil
}

func (i *UpdateUserInput) Validate() error {
	if i.Username != nil && *i.Username == "" {
		return ErrInvalidUsername
	}
	if i.Email != nil && *i.Email == "" {
		return ErrInvalidEmail
	}
	if i.Password != nil && *i.Password == "" {
		return ErrInvalidPassword
	}
	return nil
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}
