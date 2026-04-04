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

type TokenManager interface {
	GenerateToken(userID uuid.UUID, username string) (string, error)
	Validate(tokenStr string) (uuid.UUID, error)
}
