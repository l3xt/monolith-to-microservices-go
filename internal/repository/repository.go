package repository

import (
	"bookshelf/internal/database"
)

type Repository struct {
	User   *UserRepository
	Book   *BookRepository
	Review *ReviewRepository
}

func New(db *database.PostgresDB) *Repository {
	if db == nil {
		panic("db is nil")
	}
	return &Repository{
		User:   NewUserRepository(db),
		Book:   NewBookRepository(db),
		Review: NewReviewRepository(db),
	}
}
