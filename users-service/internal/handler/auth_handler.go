package handler

import (
	"bookshelf/users-service/internal/domain"
	applogger "bookshelf/users-service/internal/logger"
	"bookshelf/users-service/internal/transport/http/dto"
	"bookshelf/users-service/internal/transport/http/mapper"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
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
	Logout(ctx context.Context, token string) error
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
		case errors.Is(err, domain.ErrEmptyEmail):
			errDetails = append(errDetails, invalidEmail)
		case errors.Is(err, domain.ErrEmptyUsername):
			errDetails = append(errDetails, invalidUsername)
		case errors.Is(err, domain.ErrEmptyPassword):
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
		TokenType: "Bearer",
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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Если токена нет, считаем, что пользователь уже разлогинен
		writeJSON(w, http.StatusOK, map[string]string{"message": "successfully logged out"})
		return
	}

	var token string
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && parts[0] == "Bearer" {
		token = parts[1]
	}

	if token != "" {
		if err := h.authService.Logout(r.Context(), token); err != nil {
			log.Error("failed to logout user", slog.Any("error", err))
			writeSystemError(w, r, "Failed to logout")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "successfully logged out"})
}
