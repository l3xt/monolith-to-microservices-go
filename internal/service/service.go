package service

import (
	"errors"
)

var (
	ErrEmptyID           = errors.New("empty id value")
	ErrInvalidInputValue = errors.New("invalid input value")
)
