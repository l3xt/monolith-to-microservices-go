package dto

import (
	"github.com/google/uuid"
)

type TokenRequest struct {
	Token string `json:"token"`
}

type VerifyResponse struct {
	Valid     bool      `json:"valid"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt int64     `json:"expires_at"`
	Error     string    `json:"error"`
}

type GetUsersRequest struct {
	UserIDs []uuid.UUID `json:"ids"`
}

type GetUsersResponse struct {
	Users []UserPublic `json:"users"`
}
