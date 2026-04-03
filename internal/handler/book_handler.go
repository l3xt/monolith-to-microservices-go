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
	invalidBookTitle  = dto.ErrorDetail{Field: "Title", Message: "Invalid book title value"}
	invalidBookAuthor = dto.ErrorDetail{Field: "Author", Message: "Invalid book author value"}
)

type BookUseCase interface {
	List(ctx context.Context, filter *domain.BookFilter) ([]domain.Book, map[uuid.UUID]*domain.User, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, *domain.User, error)
	Create(ctx context.Context, userID uuid.UUID, input *domain.CreateBookInput) (*domain.Book, *domain.User, error)
	Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input *domain.UpdateBookInput) (*domain.Book, *domain.User, error)
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}

type BookHandler struct {
	bookService BookUseCase
}

func NewBookHandler(bs BookUseCase) *BookHandler {
	return &BookHandler{bookService: bs}
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	page, err := getIntParam(r, "page")
	if err != nil && !errors.Is(err, ErrEmptyParam) {
		log.Warn("invalid page param", slog.Any("error", err))
	}

	limit, err := getIntParam(r, "limit")
	if err != nil && !errors.Is(err, ErrEmptyParam) {
		log.Warn("invalid limit param", slog.Any("error", err))
	}

	search := r.URL.Query().Get("search")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	filter := domain.BookFilter{
		Page:   page,
		Limit:  limit,
		Search: search,
		Sort:   sort,
		Order:  order,
	}
	books, usersMap, totalCount, err := h.bookService.List(r.Context(), &filter)
	if err != nil {
		log.Error("failed to get book list", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get book list")
		return
	}

	responses := make([]dto.BookResponse, 0, len(books))
	for _, b := range books {
		uSummary := mapper.ToUserSummary(usersMap[b.CreatedBy])
		responses = append(responses, *mapper.ToBookResponse(&b, uSummary))
	}

	resp := &dto.BookListResponse{
		Data:       responses,
		Pagination: dto.NewPagination(page, limit, totalCount),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	rawBookID := chi.URLParam(r, "bookId")
	bookID, err := uuid.Parse(rawBookID)
	if err != nil {
		log.Warn("invalid bookId param", slog.String("raw_book_id", rawBookID))
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
		return
	}
	log = log.With(slog.String("book_id", bookID.String()))

	book, user, err := h.bookService.GetByID(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			log.Warn("failed to get book by id", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
			return
		}

		log.Error("failed to get book by id", slog.Any("error", err))
		writeSystemError(w, r, "Failed to get book")
		return
	}

	resp := &dto.BookResponse{
		ID:            book.ID,
		Title:         book.Title,
		Author:        book.Author,
		Description:   book.Description,
		ISBN:          book.ISBN,
		PublisherYear: book.PublishedYear,
		AverageRating: book.AverageRating,
		ReviewsCount:  book.ReviewsCount,
		CreatedBy:     mapper.ToUserSummary(user),
		CreatedAt:     book.CreatedAt,
		UpdatedAt:     book.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	log := applogger.FromContext(r.Context())

	userID, err := getUserID(r.Context())
	if err != nil {
		log.Warn("failed to get userID in context", slog.Any("error", err))
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "User is unauthorized", nil)
		return
	}

	req, err := decodeJSON[dto.CreateBookRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode book create request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode book create request")
		return
	}

	var errDetails []dto.ErrorDetail
	if req.Title == "" {
		errDetails = append(errDetails, invalidBookTitle)
	}
	if req.Author == "" {
		errDetails = append(errDetails, invalidBookAuthor)
	}
	if len(errDetails) > 0 {
		writeValidationError(w, r, errDetails)
		return
	}

	in := &domain.CreateBookInput{
		Title:         req.Title,
		Author:        req.Author,
		Description:   req.Description,
		ISBN:          req.ISBN,
		PublishedYear: req.PublishedYear,
	}

	book, user, err := h.bookService.Create(r.Context(), userID, in)
	if err != nil {
		log.Error("failed to create book", slog.Any("error", err))
		writeSystemError(w, r, "Failed to create book")
		return
	}

	writeJSON(w, http.StatusCreated, mapper.ToBookResponse(book, mapper.ToUserSummary(user)))
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
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

	req, err := decodeJSON[dto.UpdateBookRequest](w, r, defaultMaxBytes)
	if err != nil {
		log.Error("failed to decode book update request", slog.Any("error", err))
		writeSystemError(w, r, "Failed to decode book update request")
		return
	}

	in := &domain.UpdateBookInput{
		Title:         req.Title,
		Author:        req.Author,
		Description:   req.Description,
		ISBN:          req.ISBN,
		PublishedYear: req.PublishedYear,
	}

	book, user, err := h.bookService.Update(r.Context(), userID, bookID, in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotBookOwner):
			log.Warn("no rights to update book", slog.Any("error", err))
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "No rights to update book", nil)
		case errors.Is(err, service.ErrBookNotFound):
			log.Warn("failed to update book", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
		default:
			log.Error("failed to update book", slog.Any("error", err))
			writeSystemError(w, r, "Failed to create book")
		}
		return
	}

	writeJSON(w, http.StatusOK, mapper.ToBookResponse(book, mapper.ToUserSummary(user)))
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
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

	if err := h.bookService.Delete(r.Context(), userID, bookID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotBookOwner):
			log.Warn("user is not book owner", slog.Any("error", err))
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Not book owner", nil)
		case errors.Is(err, service.ErrBookNotFound):
			log.Warn("failed to delete book", slog.Any("error", err))
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Book not found", nil)
		default:
			log.Error("failed to delete book", slog.Any("error", err))
			writeSystemError(w, r, "Failed to create book")
		}
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
