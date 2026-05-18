package handler

import (
	"bookshelf/users-service/internal/domain"
	applogger "bookshelf/users-service/internal/logger"
	"bookshelf/users-service/internal/service"
	"bookshelf/users-service/internal/transport/http/dto"
	"bookshelf/users-service/internal/transport/http/mapper"
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type InternalUserUseCase interface {
	GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]domain.User, error)
}

type InternalHandler struct {
	tokenManager service.TokenManager
	userService  InternalUserUseCase
}

func NewInternalHandler(tm service.TokenManager, us InternalUserUseCase) *InternalHandler {
	return &InternalHandler{
		tokenManager: tm,
		userService:  us,
	}

}

func (h *InternalHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	req, err := decodeJSON[dto.TokenRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode token request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode token request")
		return
	}

	claims, err := h.tokenManager.Validate(req.Token)
	if err != nil {
		log.Error("failed to verify token", slog.Any("error", err))
		writeJSON(w, http.StatusOK, &dto.VerifyResponse{
			Valid: false,
			Error: "token invalid",
		})
		return
	}

	writeJSON(w, http.StatusOK, &dto.VerifyResponse{
		Valid:     true,
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt,
		Error:     "",
	})
}

func (h *InternalHandler) GetUsersByIDs(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	req, err := decodeJSON[dto.GetUsersRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode get public users request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode get public users request")
		return
	}

	users, err := h.userService.GetByIDs(r.Context(), req.UserIDs)
	if err != nil {
		log.Error("failed to get public users", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get public users")
		return
	}

	publicUsers := make([]dto.UserPublic, 0, len(users))
	for _, u := range users {
		publicUsers = append(publicUsers, *mapper.ToUserPublic(&u))
	}

	resp := dto.GetUsersResponse{
		Users: publicUsers,
	}

	writeJSON(w, http.StatusOK, resp)
}
