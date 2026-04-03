package handler

import (
	"bookshelf/internal/domain"
	applogger "bookshelf/internal/logger"
	"bookshelf/internal/service"
	"bookshelf/internal/transport/http/dto"
	"bookshelf/internal/transport/http/mapper"
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
	ListByBookID(ctx context.Context, bookID uuid.UUID, page, limit int) ([]domain.Review, map[uuid.UUID]*domain.User, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, *domain.User, error)
	Create(ctx context.Context, userID uuid.UUID, bookID uuid.UUID, input domain.CreateReviewInput) (*domain.Review, *domain.User, error)
	Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input *domain.UpdateReviewInput) (*domain.Review, *domain.User, error)
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

	reviews, authorsMap, totalCount, err := h.reviewService.ListByBookID(r.Context(), bookID, page, limit)
	if err != nil {
		log.Error("failed to get review list", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get review list")
		return
	}

	// Заполняем срез с ReviewResponse
	responses := make([]dto.ReviewResponse, 0, totalCount)
	for _, r := range reviews {
		uSummary := mapper.ToUserSummary(authorsMap[r.UserID])
		responses = append(responses, *mapper.ToReviewResponse(&r, uSummary))
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

	review, author, err := h.reviewService.GetByID(r.Context(), reviewID)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			log.Warn("failed to get review by id", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Review not found", nil)
			return
		}

		log.Error("failed to get review by id", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get review")
		return
	}

	resp := &dto.ReviewResponse{
		ID:        review.ID,
		BookID:    review.BookID,
		User:      mapper.ToUserSummary(author),
		Rating:    review.Rating,
		Title:     review.Title,
		Content:   review.Content,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
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
	if req.Rating < service.MinRating || req.Rating > service.MaxRating {
		errDetails = append(errDetails, invalidReviewRating)
	}
	if req.Content == "" {
		errDetails = append(errDetails, invalidReviewContent)
	}

	if len(errDetails) > 0 {
		writeValidationError(w, r, errDetails)
		return
	}

	in := domain.CreateReviewInput{
		Rating:  req.Rating,
		Title:   req.Title,
		Content: req.Content,
	}

	review, user, err := h.reviewService.Create(r.Context(), userID, bookID, in)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyReviewed) {
			log.Info("failed recreate a review from the user", slog.Any("error", err))
			writeError(w, r, http.StatusConflict, "CONFLICT", "A user has already left a review for this book", nil)
			return
		}

		log.Error("failed to create review", slog.Any("error", err))
		writeSystemError(w, r, "Failed to create review")
		return
	}

	resp := &dto.ReviewResponse{
		ID:        review.ID,
		BookID:    review.BookID,
		User:      mapper.ToUserSummary(user),
		Rating:    review.Rating,
		Title:     review.Title,
		Content:   review.Content,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
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

	in := &domain.UpdateReviewInput{
		Rating:  req.Rating,
		Title:   req.Title,
		Content: req.Content,
	}

	review, author, err := h.reviewService.Update(r.Context(), userID, reviewID, in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReviewContentTooShort):
			writeValidationError(w, r, []dto.ErrorDetail{invalidReviewContent})
		case errors.Is(err, service.ErrInvalidRating):
			writeValidationError(w, r, []dto.ErrorDetail{invalidReviewRating})
		case errors.Is(err, service.ErrNotReviewOwner):
			log.Warn("no rights to update review", slog.Any("error", err))
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "No rights to update review", nil)
		case errors.Is(err, service.ErrReviewNotFound):
			log.Warn("failed to update review", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Review not found", nil)
		default:
			log.Error("failed to update review", slog.Any("error", err))
			writeSystemError(w, r, "Failed to update review")
		}
		return
	}

	resp := &dto.ReviewResponse{
		ID:        review.ID,
		BookID:    review.BookID,
		User:      mapper.ToUserSummary(author),
		Rating:    review.Rating,
		Title:     review.Title,
		Content:   review.Content,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
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
		case errors.Is(err, service.ErrNotReviewOwner):
			log.Warn("no rights to delete review", slog.Any("error", err))
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "No rights to delete review", nil)
		case errors.Is(err, service.ErrReviewNotFound):
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
