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

	"github.com/google/uuid"
)

type UserUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, input *domain.UpdateUserInput) (*domain.User, error)
}

type UserHandler struct {
	userService UserUseCase
}

func NewUserHandler(us UserUseCase) *UserHandler {
	return &UserHandler{userService: us}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	id, err := getUserID(r.Context())
	if err != nil {
		log.Warn("failed to get userID in context", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get user id from context")
		return
	}
	
	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		log.Error("failed to get current user", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get current user")
		return
	}

	writeJSON(w, http.StatusOK, mapper.ToUserPublic(user))
}


func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	id, err := getUserID(r.Context())
	if err != nil {
		log.Warn("failed to get userID in context", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get user id from context")
		return
	}

	req, err := decodeJSON[dto.UpdateUserRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode update request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode update request")
		return
	}

	input := &domain.UpdateUserInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	var errDetails []dto.ErrorDetail
	user, err := h.userService.Update(r.Context(), id, input)
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

		log.Error("failed to update current user", slog.Any("error", err))
		writeSystemError(w, r, "Failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, mapper.ToUserPublic(user))
}
