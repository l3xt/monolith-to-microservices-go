package domain

import "errors"

// Глобальные бизнес-ошибки, которые считаются "постоянными" (нет смысла повторять)
var (
	ErrInvalidFormat = errors.New("invalid image format")
	ErrImageTooLarge = errors.New("image is too large")
)

// FatalError - кастомный тип ошибки-обертки для любых непредвиденных фатальных ошибок
type FatalError struct {
	Err error
}

func (e *FatalError) Error() string {
	return e.Err.Error()
}

// Unwrap позволяет использовать errors.Is и errors.As для вложенной ошибки
func (e *FatalError) Unwrap() error {
	return e.Err
}

// NewFatalError оборачивает ошибку, чтобы консьюмер понял, что ее не нужно возвращать в очередь
func NewFatalError(err error) error {
	if err == nil {
		return nil
	}
	return &FatalError{Err: err}
}
