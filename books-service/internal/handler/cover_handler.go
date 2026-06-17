package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"bookshelf/books-service/internal/domain"
	applogger "bookshelf/books-service/internal/logger"
	"bookshelf/books-service/internal/transport/http/dto"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	coverMaxBytes = int64(5 << 20) // 5mb
)

type CoverUseCase interface {
	UploadBookCover(ctx context.Context, userID, bookID uuid.UUID, reader io.Reader, filename, bucketName string, size int64) (uuid.UUID, error)
}

type CoverHandler struct {
	bucketName   string
	coverService CoverUseCase
}

func NewCoverHandler(useCase CoverUseCase, bucketName string) *CoverHandler {
	return &CoverHandler{
		coverService: useCase,
		bucketName:   bucketName,
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

	coverID, err := h.coverService.UploadBookCover(r.Context(), userID, bookID, file, header.Filename, h.bucketName, header.Size)
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
