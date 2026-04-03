package dto

import (
	"time"

	"github.com/google/uuid"
)

// Для передачи
type UserPublic struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Для страивания
type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

// REQUEST/RESPONSE BLOCK
// Регистрация юзера
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Авторизация юзера
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Доступ автарику
type AuthResponse struct {
	User         *UserPublic `json:"user"`
	AccessToken  string      `json:"access_token"`
}

// Обновление юзера
type UpdateUserRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

