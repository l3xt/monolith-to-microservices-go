package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/logger"
	applogger "bookshelf/books-service/internal/logger"
	"bookshelf/books-service/internal/transport/http/dto"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	coverMaxBytes = int64(5 << 20) // 5mb
)

type CoverUseCase interface {
	UploadBookCover(ctx context.Context, userID, bookID uuid.UUID, reader io.Reader, size int64) (uuid.UUID, error)
	GetCover(ctx context.Context, bookID uuid.UUID) (*domain.Cover, error)
	DeleteCover(ctx context.Context, userID, bookID uuid.UUID) error
}

type CoverHandler struct {
	coverService CoverUseCase
}

func NewCoverHandler(useCase CoverUseCase) *CoverHandler {
	return &CoverHandler{
		coverService: useCase,
	}
}

func (h *CoverHandler) UploadBookCover(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, coverMaxBytes)

	if err := r.ParseMultipartForm(coverMaxBytes); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			log.Warn("cover exceeded max size", slog.Any("error", err))
			writeError(w, r, http.StatusBadRequest, "FILE_TOO_LARGE", "File size exceeds maximum allowed (5MB)", nil)
			return
		}

		log.Error("failed to parse multipart form", slog.Any("error", err))
		writeSystemError(w, r, "Failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Error("failed to get file", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get file")
		return
	}

	defer file.Close()

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

	coverID, err := h.coverService.UploadBookCover(r.Context(), userID, bookID, file, header.Size)
	if err != nil {
		log.Error("failed to upload cover", slog.Any("error", err))
		if errors.Is(err, domain.ErrNotBookOwner) {
			log.Warn("no rights to upload cover", slog.Any("error", err))
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "No rights to upload cover", nil)
			return
		}
		if errors.Is(err, domain.ErrBookNotFound) {
			log.Warn("failed to upload cover", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
			return
		}

		if errors.Is(err, domain.ErrInvalidCoverType) {
			log.Warn("invalid cover type", slog.Any("error", err))
			writeError(w, r, http.StatusBadRequest, "INVALID_FILE_TYPE", "Only jpg, jpeg, png, webp files are allowed", nil)
			return
		}
		if errors.Is(err, domain.ErrCoverExceededSize) {
			log.Warn("cover exceeded max size", slog.Any("error", err))
			writeError(w, r, http.StatusBadRequest, "FILE_TOO_LARGE", "File size exceeds maximum allowed (5MB)", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidCoverFile) {
			log.Warn("invalid cover file", slog.Any("error", err))
			writeError(w, r, http.StatusBadRequest, "INVALID_FILE_TYPE", "Invalid cover file", nil)
			return
		}

		writeSystemError(w, r, "Failed to upload cover")
		return
	}

	resp := dto.CoverUploadResponse{
		CoverID: coverID,
		Status:  domain.CoverStatusProcessing,
		Message: "Cover upload succefully",
	}

	writeJSON(w, http.StatusAccepted, resp)
}

func (h *CoverHandler) GetBookCover(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	
	rawBookID := chi.URLParam(r, "bookId")
	bookID, err := uuid.Parse(rawBookID)
	if err != nil {
		log.Warn("invalid bookId param", slog.String("raw_book_id", rawBookID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
		return
	}
	log = log.With(slog.String("book_id", bookID.String()))

	cover, err := h.coverService.GetCover(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, domain.ErrCoverNotFound) {
			log.Warn("cover not found", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Cover not found", nil)
			return
		}
		log.Error("failed to get cover", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get cover")
		return
	}

    var coverURL, thumbURL, message *string
    switch cover.Status {
    case domain.CoverStatusReady:
        coverURL = cover.CoverURL
        thumbURL = cover.ThumbURL
    case domain.CoverStatusProcessing:
        msg := "Processing..."
        message = &msg
    case domain.CoverStatusFailed:
        message = cover.Error
    case domain.CoverStatusNone:
        msg := "No cover"
        message = &msg
    }
    resp := dto.CoverResponse{
        Status:   cover.Status,
        CoverURL: coverURL,
        ThumbURL: thumbURL,
        Message:  message,
    }
    writeJSON(w, http.StatusOK, resp)
}

func (h *CoverHandler) GetBookCoverStatus(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	
	rawBookID := chi.URLParam(r, "bookId")
	bookID, err := uuid.Parse(rawBookID)
	if err != nil {
		log.Warn("invalid bookId param", slog.String("raw_book_id", rawBookID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
		return
	}
	log = log.With(slog.String("book_id", bookID.String()))

	cover, err := h.coverService.GetCover(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, domain.ErrCoverNotFound) {
			log.Warn("cover not found", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Cover not found", nil)
			return
		}
		log.Error("failed to get cover", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get cover")
		return
	}

    resp := dto.CoverStatusResponse{
		Status: cover.Status,
		CoverURL: cover.CoverURL,
		ThumbURL: cover.ThumbURL,
		Error: cover.Error,
		CreatedAt: cover.CreatedAt,
		CompletedAt: cover.CompletedAt,
    }
    writeJSON(w, http.StatusOK, resp)
}

func (h *CoverHandler) DeleteBookCover(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	
	rawBookID := chi.URLParam(r, "bookId")
	bookID, err := uuid.Parse(rawBookID)
	if err != nil {
		log.Warn("invalid bookId param", slog.String("raw_book_id", rawBookID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
		return
	}
	log = log.With(slog.String("book_id", bookID.String()))

	userID, err := getUserID(r.Context())
	if err != nil {
		log.Warn("failed to get userID in context", slog.Any("error", err))
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "User is unauthorized", nil)
		return
	}
	log = log.With(slog.String("user_id", userID.String()))

	if err := h.coverService.DeleteCover(r.Context(), userID, bookID); err != nil {
		if errors.Is(err, domain.ErrNotBookOwner) {
			log.Warn("no rights to delete cover", slog.Any("error", err))
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "No rights to delete cover", nil)
			return
		}
		if errors.Is(err, domain.ErrBookNotFound) {
			log.Warn("failed to delete cover", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
			return
		}
		log.Error("failed to delete cover", slog.Any("error", err))
		writeSystemError(w, r, "Failed to delete cover")
		return
	}

    writeJSON(w, http.StatusOK, nil)
}
