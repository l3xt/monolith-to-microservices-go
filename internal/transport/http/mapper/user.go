package mapper

import (
	"bookshelf/internal/domain"
	"bookshelf/internal/transport/http/dto"
)

// Конвертация в публичную модель
func ToUserPublic(u *domain.User) *dto.UserPublic {
	return &dto.UserPublic{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// Конвертация в краткую модель
func ToUserSummary(u *domain.User) *dto.UserSummary {
	return &dto.UserSummary{
		ID: u.ID,
		Username: u.Username,
	}
}