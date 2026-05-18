package service

import (
	"bookshelf/books-service/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrEmptyBookID = errors.New("empty id value")
)

type BookService struct {
	bookRepo domain.BookRepository
}

func NewBookService(b domain.BookRepository) *BookService {
	return &BookService{
		bookRepo: b,
	}
}

func (s *BookService) Create(ctx context.Context, userID uuid.UUID, input *domain.CreateBookInput) (*domain.Book, error) {
	// Валидация
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("BookService.Create: validate input: %w", err)
	}

	// Создание книги
	book := &domain.Book{
		Title:         input.Title,
		Author:        input.Author,
		Description:   input.Description,
		ISBN:          input.ISBN,
		PublishedYear: input.PublishedYear,
		UserID:        userID,
	}

	err := s.bookRepo.Create(ctx, book)
	if err != nil {
		return nil, fmt.Errorf("BookService.Create: create book: %w", err)
	}

	return book, nil
}

func (s *BookService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	// Валидация
	if id == uuid.Nil {
		return nil, ErrEmptyBookID
	}

	// Получаем модель книги
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("BookService.GetByID: fetch book: %w", err)
	}

	return book, nil
}

func (s *BookService) List(ctx context.Context, filter *domain.BookFilter) ([]domain.Book, int, error) {
	// defauls для фильтров
	filter.SetDefaults()

	// Валидация
	if err := filter.Validate(); err != nil {
		return nil, 0, fmt.Errorf("BookService.List: filter validation: %w", err)
	}

	// Получаем список книг
	books, bCount, err := s.bookRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("BookService.List: fetch list: %w", err)
	}

	return books, bCount, nil
}

func (s *BookService) Update(ctx context.Context, userID, bookID uuid.UUID, req *domain.UpdateBookInput) (*domain.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("BookService.Update: fetch book: %w", err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Проверяем владельца книги
	if book.UserID != userID {
		return nil, domain.ErrNotBookOwner
	}

	if req.Title != nil {
		book.Title = *req.Title
	}
	if req.Author != nil {
		book.Author = *req.Author
	}
	if req.Description != nil {
		book.Description = req.Description
	}
	if req.ISBN != nil {
		book.ISBN = req.ISBN
	}
	if req.PublishedYear != nil {
		book.PublishedYear = req.PublishedYear
	}

	if err := s.bookRepo.Update(ctx, book); err != nil {
		return nil, fmt.Errorf("BookService.Update: update book: %w", err)
	}

	return book, nil
}

func (s *BookService) Delete(ctx context.Context, userID, bookID uuid.UUID) error {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return fmt.Errorf("BookService.Delete: fetch book: %w", err)
	}

	// Проверяем владельца книги
	if book.UserID != userID {
		return domain.ErrNotBookOwner
	}

	if err := s.bookRepo.Delete(ctx, book.ID); err != nil {
		return fmt.Errorf("BookService.Delete: delete book: %w", err)
	}
	return nil
}
