package service

import (
	"bookshelf/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (*domain.User, error) {
	// Хеширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создание пользователя
	newUser := domain.NewUser(username, email, string(hash))
	err = s.repo.Create(ctx, newUser)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) || errors.Is(err, domain.ErrUsernameExists) {
			return nil, err
		}
		return nil, fmt.Errorf("UserService.Register: %w", err)
	}

	return newUser, nil
}

func (s *UserService) GetUserByCredentials(ctx context.Context, email, password string) (*domain.User, error) {
	// Ищем пользователя по email
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *UserService) Update(ctx context.Context, userID uuid.UUID, input *domain.UpdateUserInput) (*domain.User, error) {
	// Ищем пользователя по ID
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Проверяем наличие username
	if input.Username != nil && user.Username != *input.Username {
		// Проверяем новый username на уникальность
		u, err := s.repo.GetByUsername(ctx, *input.Username)
		if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
			return nil, err
		}
		if u != nil {
			return nil, domain.ErrUsernameExists
		}
		user.Username = *input.Username
	}

	// Проверяем наличие email
	if input.Email != nil && user.Email != *input.Email {
		// Проверяем новый email на уникальность
		u, err := s.repo.GetByEmail(ctx, *input.Email)
		if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
			return nil, err
		}
		if u != nil {
			return nil, domain.ErrUserExists
		}
		user.Email = *input.Email
	}

	if input.Password != nil {
		// Хеширование пароля
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		user.PasswordHash = string(hash)
	}

	return user, s.repo.Update(ctx, user)
}
