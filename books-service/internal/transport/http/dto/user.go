package dto

import (
	"time"

	"github.com/google/uuid"
)

// Для страивания
type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type UserPublic struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenRequest struct {
	Token string `json:"token"`
}

type GetUsersRequest struct {
	UserIDs []uuid.UUID `json:"ids"`
}

type VerifyResponse struct {
	Valid     bool   `json:"valid"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
	Error     string `json:"error"`
}

type GetUsersResponse struct {
	Users []UserPublic `json:"users"`
}
