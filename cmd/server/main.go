package main

import (
	"bookshelf/internal/config"
	"bookshelf/internal/database"
	"bookshelf/internal/handler"
	appjwt "bookshelf/internal/lib/jwt"
	applogger "bookshelf/internal/logger"
	"bookshelf/internal/repository"
	"bookshelf/internal/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const DefaultTimeout = 30 * time.Second

func main() {
	env := "dev"
	logger := applogger.New(env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx, logger); err != nil {
		logger.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dbCfg := database.Config{
		URL:               cfg.DatabaseURL,
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: 10 * time.Second,
	}

	db, err := database.NewPostgresDB(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}
	defer db.Close()

	log.Info("connected to database")

	// Инициализация инфраструктуры
	jwtManager := appjwt.NewJWTManager(cfg.SecretKey)

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// Инициализация бизнес-логики
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userService, jwtManager)
	bookService := service.NewBookService(userRepo, bookRepo)
	reviewService := service.NewReviewService(userRepo, bookRepo, reviewRepo)

	// Инициализация транспортного слоя
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	bookHandler := handler.NewBookHandler(bookService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	systemHandler := handler.NewSystemHandler(db)


	// Инициализация роутера
	router := newRouter(log, authHandler, userHandler, bookHandler, reviewHandler, systemHandler, jwtManager)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("server started", slog.String("server_address", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info("stopping the server", slog.Any("signal", sig))

		if err := server.Shutdown(ctx); err != nil {
			log.Info("server could not be stopped, forced termination", slog.Any("error", err))
			server.Close()
		}
	}

	log.Info("server was successfully stopped")
	return nil
}

// В идеале роутер принимает конкретные интерфейсы хэндлеров и middleware,
// а не глобальный объект.
func newRouter(logger *slog.Logger, authH *handler.AuthHandler, userH *handler.UserHandler, bookH *handler.BookHandler, reviewH *handler.ReviewHandler, systemHandler *handler.SystemHandler, tokenManager service.TokenManager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// MIDDLEWARE
	r.Use(handler.LoggingMiddleware(logger))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Устанавливаем таймаут
	r.Use(middleware.Timeout(DefaultTimeout))

	// ENDPOINTS
	r.Get("/health", systemHandler.Health)
	r.Get("/ready", systemHandler.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		// Публичные
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)

		r.Get("/books", bookH.ListBooks)
		r.Get("/books/{bookId}", bookH.GetBook)

		r.Get("/books/{bookId}/reviews", reviewH.ListBookReviews)
		r.Get("/reviews/{reviewId}", reviewH.GetReview)

		// Защищенные
		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(tokenManager))

			r.Get("/users/me", userH.GetCurrentUser)
			r.Put("/users/me", userH.UpdateCurrentUser)

			r.Post("/books", bookH.CreateBook)
			r.Put("/books/{bookId}", bookH.UpdateBook)
			r.Delete("/books/{bookId}", bookH.DeleteBook)

			r.Post("/books/{bookId}/reviews", reviewH.CreateReview)
			r.Put("/reviews/{reviewId}", reviewH.UpdateReview)
			r.Delete("/reviews/{reviewId}", reviewH.DeleteReview)
		})
	})

	return r
}
