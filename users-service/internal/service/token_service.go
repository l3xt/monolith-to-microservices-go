package service

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrTokenExpired     = errors.New("expired token")
	ErrTokenNotValidYet = errors.New("token not yet valid")
	ErrTokenMalformed   = errors.New("malformed token")
)

// UserClaims содержит полезную нагрузку (payload) из токена
type UserClaims struct {
	UserID   uuid.UUID
	Username string
	ExpiresAt int64
}

type TokenManager interface {
	GenerateToken(userID uuid.UUID, username string) (string, error)
	Validate(tokenStr string) (*UserClaims, error)
}
