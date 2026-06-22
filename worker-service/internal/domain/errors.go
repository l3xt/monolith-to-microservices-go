package domain

import "errors"

// Бизнес ошибки
var (
	// Книги
	ErrBookNotFound = errors.New("book not found")

	// Обложки
	ErrCoverNotFound     = errors.New("cover not found")
	ErrInvalidCoverFile  = errors.New("invalid cover file")
	ErrCoverExceededSize = errors.New("cover exceeded max size")
	ErrInvalidCoverType  = errors.New("invalid cover type")

	// Остальное
	ErrInvalidFormat = errors.New("invalid image format")
	ErrImageTooLarge = errors.New("image is too large")
	ErrServiceNotResponding = errors.New("service is not responding")
)
