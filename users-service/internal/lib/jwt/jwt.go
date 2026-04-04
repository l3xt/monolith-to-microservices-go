package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidTokenClaims = errors.New("invalid token claims")

type JWTManager struct {
	secretKey []byte
}

type Claims struct {
	UserID   uuid.UUID `json:"uid"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret []byte) *JWTManager {
	return &JWTManager{secretKey: secret}
}

func (m *JWTManager) GenerateToken(userID uuid.UUID, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "bookshelf",
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// Валидация токена
func (m *JWTManager) Validate(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			// проверка алгоритма
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return uuid.Nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secretKey, nil
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	// Распаковываем интерфейс в нашу структуру
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidTokenClaims
	}

	return claims.UserID, nil
}
