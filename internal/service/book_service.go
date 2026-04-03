package service

import (
	"bookshelf/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	DefaultBookLimit = 100
	DefaultBookSort  = "title"
)

var (
	ErrBookNotFound    = errors.New("book not found")
	ErrNotBookOwner    = errors.New("not book owner")
	ErrBookTitleEmpty  = errors.New("empty book title")
	ErrBookAuthorEmpty = errors.New("empty book author")
)

type BookService struct {
	bookRepo domain.BookRepository
	userRepo domain.UserRepository
}

func NewBookService(u domain.UserRepository, b domain.BookRepository) *BookService {
	return &BookService{
		userRepo: u,
		bookRepo: b,
	}
}

func (s *BookService) Create(ctx context.Context, userID uuid.UUID, in *domain.CreateBookInput) (*domain.Book, *domain.User, error) {
	// Валидация
	if in.Title == "" {
		return nil, nil, ErrBookTitleEmpty
	}
	if in.Author == "" {
		return nil, nil, ErrBookAuthorEmpty
	}

	// Получаем модель автора книги
	creator, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch creator: %w", err)
	}

	// Создание книги
	book := domain.Book{
		Title:         in.Title,
		Author:        in.Author,
		Description:   in.Description,
		ISBN:          in.ISBN,
		PublishedYear: in.PublishedYear,
		CreatedBy:     creator.ID,
		AverageRating: 0,
	}
	err = s.bookRepo.Create(ctx, &book)
	if err != nil {
		return nil, nil, fmt.Errorf("create book: %w", err)
	}

	return &book, creator, nil
}

func (s *BookService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, *domain.User, error) {
	if id == uuid.Nil {
		return nil, nil, ErrEmptyID
	}

	// Получаем модель книги
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			return nil, nil, ErrBookNotFound
		}
		return nil, nil, fmt.Errorf("fetch book: %w", err)
	}

	// Получаем модель автора книги
	creator, err := s.userRepo.GetByID(ctx, book.CreatedBy)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch creator: %w", err)
	}

	return book, creator, nil
}

func (s *BookService) List(ctx context.Context, filter *domain.BookFilter) ([]domain.Book, map[uuid.UUID]*domain.User, int, error) {
	// defauls для фильтров
	if filter == nil {
		filter = &domain.BookFilter{}
	}
	if filter.Limit <= 0 {
		filter.Limit = DefaultBookLimit
	}
	if filter.Sort == "" {
		filter.Sort = DefaultBookSort
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	books, bCount, err := s.bookRepo.List(ctx, filter)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("fetch book list: %w", err)
	}
	if len(books) == 0 {
		return []domain.Book{}, nil, 0, nil
	}

	// Собираем уникальных авторов
	creatorIDs := make([]uuid.UUID, 0, bCount)
	seen := make(map[uuid.UUID]struct{})
	for _, b := range books {
		if _, ok := seen[b.CreatedBy]; !ok {
			creatorIDs = append(creatorIDs, b.CreatedBy)
			seen[b.CreatedBy] = struct{}{}
		}
	}

	// Получаем данные об этих авторах
	users, err := s.userRepo.GetByIDs(ctx, creatorIDs)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("BookService.List: %w", err)
	}

	// Создаем мапу для O(1) доступа
	usersMap := make(map[uuid.UUID]*domain.User, len(users))
	for _, u := range users {
		usersMap[u.ID] = &u
	}

	return books, usersMap, bCount, nil
}

func (s *BookService) Update(ctx context.Context, userID, bookID uuid.UUID, req *domain.UpdateBookInput) (*domain.Book, *domain.User, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, nil, fmt.Errorf("BookService.Update: %w", err)
	}

	if req.Title != nil && *req.Title != "" {
		book.Title = *req.Title
	}
	if req.Author != nil && *req.Author != "" {
		book.Author = *req.Author
	}
	if req.Description != nil && *req.Description != "" {
		book.Description = req.Description
	}
	if req.ISBN != nil && *req.ISBN != "" {
		book.ISBN = req.ISBN
	}
	if req.PublishedYear != nil && *req.PublishedYear > 0 {
		book.PublishedYear = req.PublishedYear
	}

	// Проверяем владельца книги
	if book.CreatedBy != userID {
		return nil, nil, ErrNotBookOwner
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("BookService.Update: %w", err)
	}

	if err := s.bookRepo.Update(ctx, book); err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("BookService.Update: %w", err)
	}

	return book, user, nil
}

func (s *BookService) Delete(ctx context.Context, userID, bookID uuid.UUID) error {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			return err
		}
		return fmt.Errorf("BookService.Delete: %w", err)
	}

	// Проверяем владельца книги
	if book.CreatedBy != userID {
		return ErrNotBookOwner
	}

	if err := s.bookRepo.Delete(ctx, book.ID); err != nil {
		return fmt.Errorf("BookService.Delete: %w", err)
	}
	return nil
}
