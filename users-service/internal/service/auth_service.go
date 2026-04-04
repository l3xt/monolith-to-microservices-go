package service

import (
	"bookshelf/users-service/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrInvalidInputValue = errors.New("invalid input value")

type TokenGenerator interface {
	GenerateToken(userID uuid.UUID, username string) (string, error)
}

type AuthService struct {
	userService    *UserService
	tokenGenerator TokenGenerator
}

func NewAuthService(u *UserService, g TokenGenerator) *AuthService {
	return &AuthService{
		userService:    u,
		tokenGenerator: g,
	}
}

func (s *AuthService) Register(ctx context.Context, input *domain.RegisterInput) (*domain.User, string, error) {
	if input == nil {
		return nil, "", ErrInvalidInputValue
	}

	if err := input.Validate(); err != nil {
		return nil, "", err
	}

	user, err := s.userService.CreateUser(ctx, input.Username, input.Email, input.Password)
	if err != nil {
		return nil, "", err
	}

	token, err := s.tokenGenerator.GenerateToken(user.ID, user.Username)
	return user, token, err
}

func (s *AuthService) Login(ctx context.Context, input *domain.LoginInput) (*domain.User, string, error) {
	if input == nil {
		return nil, "", ErrInvalidInputValue
	}

	if err := input.Validate(); err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	user, err := s.userService.GetUserByCredentials(ctx, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, "", err
		}
		return nil, "", fmt.Errorf("AuthService.Login: %w", err)
	}

	token, err := s.tokenGenerator.GenerateToken(user.ID, user.Username)
	return user, token, err
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return nil
}
