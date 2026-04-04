package domain

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MinUsernameLength = 3
	MinPasswordLength = 8
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrEmptyUsername      = errors.New("username cannot be empty")
	ErrShortUsername      = errors.New("username must be at least 3 characters long")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmptyEmail         = errors.New("email address can't be empty")
	ErrEmptyPassword      = errors.New("password cannot be empty")
	ErrWeakPassword       = errors.New("password must be longer than 8 characters")
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
		return ErrEmptyUsername
	}
	if utf8.RuneCountInString(i.Username) < MinUsernameLength {
		return ErrShortUsername
	}
	if i.Email == "" {
		return ErrEmptyEmail
	}
	if i.Password == "" {
		return ErrEmptyPassword
	}
	if utf8.RuneCountInString(i.Password) < MinPasswordLength {
		return ErrWeakPassword
	}
	return nil
}

func (i *LoginInput) Validate() error {
	if i.Email == "" {
		return ErrInvalidCredentials
	}
	if i.Password == "" || utf8.RuneCountInString(i.Password) < MinPasswordLength {
		return ErrInvalidCredentials
	}

	return nil
}

func (i *UpdateUserInput) Validate() error {
	if i.Username != nil && *i.Username == "" {
		return ErrEmptyUsername
	}
	if i.Username != nil && utf8.RuneCountInString(*i.Username) < MinUsernameLength {
		return ErrShortUsername
	}
	if i.Email != nil && *i.Email == "" {
		return ErrEmptyEmail
	}
	if i.Password != nil && *i.Password == "" {
		return ErrEmptyPassword
	}
	if i.Password != nil && utf8.RuneCountInString(*i.Password) < MinPasswordLength {
		return ErrWeakPassword
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
