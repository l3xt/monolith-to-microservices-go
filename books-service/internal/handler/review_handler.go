package handler

import (
	"bookshelf/books-service/internal/domain"
	applogger "bookshelf/books-service/internal/logger"
	"bookshelf/books-service/internal/transport/http/dto"
	"bookshelf/books-service/internal/transport/http/mapper"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var (
	invalidReviewRating  = dto.ErrorDetail{Field: "Rating", Message: "Rating value is not in the range from 1 to 5"}
	invalidReviewTitle   = dto.ErrorDetail{Field: "Title", Message: "Invalid review title value"}
	invalidReviewContent = dto.ErrorDetail{Field: "Content", Message: "Content length must be more than 10 characters"}
)

type ReviewUseCase interface {
	ListByBook(ctx context.Context, bookID uuid.UUID, page, limit int) ([]domain.Review, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	Create(ctx context.Context, userID uuid.UUID, bookID uuid.UUID, input domain.CreateReviewInput) (*domain.Review, error)
	Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input *domain.UpdateReviewInput) (*domain.Review, error)
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}

type ReviewHandler struct {
	reviewService ReviewUseCase
}

func NewReviewHandler(rs ReviewUseCase) *ReviewHandler {
	return &ReviewHandler{reviewService: rs}
}

// Список отзывов (Публичный)
func (h *ReviewHandler) ListBookReviews(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	rawBookID := chi.URLParam(r, "bookId")
	bookID, err := uuid.Parse(rawBookID)
	if err != nil || bookID == uuid.Nil {
		log.Warn("invalid bookId param", slog.String("raw_book_id", rawBookID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
		return
	}
	log = log.With(slog.String("book_id", bookID.String()))

	page, err := getIntParam(r, "page")
	if err != nil && !errors.Is(err, ErrEmptyParam) {
		log.Warn("invalid page param", slog.Any("error", err))
	}

	limit, err := getIntParam(r, "limit")
	if err != nil && !errors.Is(err, ErrEmptyParam) {
		log.Warn("invalid limit param", slog.Any("error", err))
	}

	reviews, totalCount, err := h.reviewService.ListByBook(r.Context(), bookID, page, limit)
	if err != nil {
		log.Error("failed to get review list", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get review list")
		return
	}

	// Заполняем срез с ReviewResponse
	responses := make([]dto.ReviewResponse, 0, totalCount)
	for _, r := range reviews {
		responses = append(responses, *mapper.ToReviewResponse(&r))
	}

	resp := &dto.ReviewListResponse{
		Data:       responses,
		Pagination: dto.NewPagination(page, limit, totalCount),
	}

	writeJSON(w, http.StatusOK, resp)
}

// Отзыв (Публичный)
func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	rawReviewID := chi.URLParam(r, "reviewId")
	reviewID, err := uuid.Parse(rawReviewID)
	if err != nil || reviewID == uuid.Nil {
		log.Warn("invalid reviewId param", slog.String("raw_review_id", rawReviewID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Review not found", nil)
		return
	}
	log = log.With(slog.String("review_id", reviewID.String()))

	review, err := h.reviewService.GetByID(r.Context(), reviewID)
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			log.Warn("failed to get review by id", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Review not found", nil)
			return
		}

		log.Error("failed to get review by id", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get review")
		return
	}

	writeJSON(w, http.StatusOK, mapper.ToReviewResponse(review))
}

// Создать отзыв (Закрытый)
func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	userID, err := getUserID(r.Context())
	if err != nil {
		log.Warn("failed to get userID in context", slog.Any("error", err))
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "User is unauthorized", nil)
		return
	}

	rawBookID := chi.URLParam(r, "bookId")
	bookID, err := uuid.Parse(rawBookID)
	if err != nil {
		log.Warn("invalid bookId param", slog.String("raw_book_id", rawBookID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
		return
	}
	log = log.With(slog.String("book_id", bookID.String()))

	req, err := decodeJSON[dto.CreateReviewRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode review create request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode review create request")
		return
	}

	var errDetails []dto.ErrorDetail
	if req.Rating < domain.ReviewMinRating || req.Rating > domain.ReviewMaxRating {
		errDetails = append(errDetails, invalidReviewRating)
	}
	if req.Content == "" {
		errDetails = append(errDetails, invalidReviewContent)
	}
	if len(errDetails) > 0 {
		writeValidationError(w, r, errDetails)
		return
	}

	input := domain.CreateReviewInput{
		Rating:  req.Rating,
		Title:   req.Title,
		Content: req.Content,
	}

	review, err := h.reviewService.Create(r.Context(), userID, bookID, input)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyReviewed) {
			log.Info("failed recreate a review from the user", slog.Any("error", err))
			writeError(w, r, http.StatusConflict, "CONFLICT", "A user has already left a review for this book", nil)
			return
		}

		log.Error("failed to create review", slog.Any("error", err))
		writeSystemError(w, r, "Failed to create review")
		return
	}

	writeJSON(w, http.StatusCreated, mapper.ToReviewResponse(review))
}

// Обновить отзыв (Закрытый)
func (h *ReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	userID, err := getUserID(r.Context())
	if err != nil {
		log.Warn("failed to get userID in context", slog.Any("error", err))
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "User is unauthorized", nil)
		return
	}

	rawReviewID := chi.URLParam(r, "reviewId")
	reviewID, err := uuid.Parse(rawReviewID)
	if err != nil || reviewID == uuid.Nil {
		log.Warn("invalid reviewId param", slog.String("raw_review_id", rawReviewID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Review not found", nil)
		return
	}
	log = log.With(slog.String("review_id", reviewID.String()))

	req, err := decodeJSON[dto.UpdateReviewRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode review update request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode review update request")
		return
	}

	input := &domain.UpdateReviewInput{
		Rating:  req.Rating,
		Title:   req.Title,
		Content: req.Content,
	}

	review, err := h.reviewService.Update(r.Context(), userID, reviewID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrReviewContentTooShort):
			writeValidationError(w, r, []dto.ErrorDetail{invalidReviewContent})

		case errors.Is(err, domain.ErrInvalidRating):
			writeValidationError(w, r, []dto.ErrorDetail{invalidReviewRating})

		case errors.Is(err, domain.ErrNotReviewOwner):
			log.Warn("no rights to update review", slog.Any("error", err))
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "No rights to update review", nil)

		case errors.Is(err, domain.ErrReviewNotFound):
			log.Warn("failed to update review", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Review not found", nil)

		default:
			log.Error("failed to update review", slog.Any("error", err))
			writeSystemError(w, r, "Failed to update review")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapper.ToReviewResponse(review))
}

// Удалить отзыв (Закрытый)
func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	userID, err := getUserID(r.Context())
	if err != nil {
		log.Warn("failed to get userID in context", slog.Any("error", err))
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "User is unauthorized", nil)
		return
	}

	rawReviewID := chi.URLParam(r, "reviewId")
	reviewID, err := uuid.Parse(rawReviewID)
	if err != nil || reviewID == uuid.Nil {
		log.Warn("invalid reviewId param", slog.String("raw_review_id", rawReviewID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Review not found", nil)
		return
	}
	log = log.With(slog.String("review_id", reviewID.String()))

	if err := h.reviewService.Delete(r.Context(), userID, reviewID); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotReviewOwner):
			log.Warn("no rights to delete review", slog.Any("error", err))
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "No rights to delete review", nil)

		case errors.Is(err, domain.ErrReviewNotFound):
			log.Warn("failed to get review", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Review not found", nil)

		default:
			log.Error("failed to delete review", slog.Any("error", err))
			writeSystemError(w, r, "Failed to delete review")
		}
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
