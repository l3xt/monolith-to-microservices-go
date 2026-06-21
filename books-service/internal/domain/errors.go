package domain

import (
	"errors"
)

// Books
var (
	ErrInvalidInputValue      = errors.New("invalid input value")
	ErrInvalidBookFilterValue = errors.New("invalid book filter value")
	ErrInvalidSortValue       = errors.New("invalid sort value in filter object")
	ErrInvalidOrderValue      = errors.New("invalid order value in filter object")
	ErrBookNotFound           = errors.New("book not found")
	ErrBookTitleEmpty         = errors.New("empty book title")
	ErrBookAuthorEmpty        = errors.New("empty book author")
	ErrNotBookOwner           = errors.New("not book owner")
)

// Covers
var (
	ErrCoverNotFound     = errors.New("cover not found")
	ErrInvalidCoverFile  = errors.New("invalid cover file")
	ErrCoverExceededSize = errors.New("cover exceeded max size")
	ErrInvalidCoverType  = errors.New("invalid cover type")
)

// Reviews
var (
	ErrReviewNotFound        = errors.New("review not found")
	ErrNotReviewOwner        = errors.New("user is not review owner")
	ErrAlreadyReviewed       = errors.New("user has already left a review about this book")
	ErrInvalidRating         = errors.New("rating value is not in the range from 1 to 5")
	ErrReviewContentTooShort = errors.New("content length must be more than 10 characters")
)

// Health
var ErrAuthServiceUnavailable = errors.New("authentication service not available")
