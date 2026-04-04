package dto

import (
	"github.com/google/uuid"
)

// Для страивания
type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

