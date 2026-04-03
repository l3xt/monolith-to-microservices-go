package handler

import (
	"bookshelf/internal/domain"
	applogger "bookshelf/internal/logger"
	"bookshelf/internal/transport/http/dto"
	"bookshelf/internal/transport/http/mapper"
	"context"
	"errors"
	"log/slog"
	"net/http"
)

var (
	invalidEmail    = dto.ErrorDetail{Field: "Email", Message: "Invalid email value"}
	invalidUsername = dto.ErrorDetail{Field: "Username", Message: "Invalid username value"}
	invalidPassword = dto.ErrorDetail{Field: "Password", Message: "Empty password value"}

	userExists     = dto.ErrorDetail{Field: "Email", Message: "The user is already registered"}
	usernameExists = dto.ErrorDetail{Field: "Username", Message: "Username already taken"}
)

type AuthUseCase interface {
	Register(ctx context.Context, input *domain.RegisterInput) (*domain.User, string, error)
	Login(ctx context.Context, input *domain.LoginInput) (*domain.User, string, error)
}

type AuthHandler struct {
	authService AuthUseCase
}

func NewAuthHandler(as AuthUseCase) *AuthHandler {
	return &AuthHandler{authService: as}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	req, err := decodeJSON[dto.RegisterRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode register request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode register request")
		return
	}

	input := &domain.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}


	var errDetails []dto.ErrorDetail
	user, token, err := h.authService.Register(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			errDetails = append(errDetails, invalidEmail)
		case errors.Is(err, domain.ErrInvalidUsername):
			errDetails = append(errDetails, invalidUsername)
		case errors.Is(err, domain.ErrInvalidPassword):
			errDetails = append(errDetails, invalidPassword)

		case errors.Is(err, domain.ErrUserExists):
			errDetails = append(errDetails, userExists)
		case errors.Is(err, domain.ErrUsernameExists):
			errDetails = append(errDetails, usernameExists)
		}

		if len(errDetails) > 0 {
			writeValidationError(w, r, errDetails)
			return
		}

		log.Error("failed to register user", slog.Any("error", err))
		writeSystemError(w, r, "Failed to register user")
		return
	}

	resp := dto.AuthResponse{
		User:        mapper.ToUserPublic(user),
		AccessToken: token,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	req, err := decodeJSON[dto.LoginRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode login request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode login request")
		return
	}

	input := &domain.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := h.authService.Login(r.Context(), input)
	if err != nil {
		// Если пользователь ввел неправильные данные, то возвращаем 401
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username or password", nil)
			return
		}

		log.Error("failed to login user", slog.Any("error", err))
		writeSystemError(w, r, "Failed to login user")
		return
	}

	resp := dto.AuthResponse{
		User:        mapper.ToUserPublic(user),
		AccessToken: token,
	}

	writeJSON(w, http.StatusOK, resp)
}


