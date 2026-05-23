package main

import (
	"bookshelf/books-service/internal/client"
	"bookshelf/books-service/internal/config"
	"bookshelf/books-service/internal/database"
	"bookshelf/books-service/internal/handler"
	applogger "bookshelf/books-service/internal/logger"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bookshelf/pkg/httpclient"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

const DefaultTimeout = 30 * time.Second

func main() {
	env := "dev"
	logger := applogger.New(env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx, logger); err != nil {
		logger.Error("failed to start users service", slog.Any("error", err))
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

	// Инициализация репозиториев
	bookRepo := repository.NewBookRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// Инициализация бизнес-логики
	bookService := service.NewBookService(bookRepo)
	reviewService := service.NewReviewService(bookRepo, reviewRepo)

	baseHTTPClient := httpclient.NewClient(cfg.AuthServiceURL, 5*time.Second)
	authClient := client.NewAuthClient(baseHTTPClient, cfg.ServiceKey)

	// Инициализация транспортного слоя
	bookHandler := handler.NewBookHandler(bookService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	systemHandler := handler.NewSystemHandler(cfg.Version, db, authClient)

	router := newRouter(bookHandler, reviewHandler, systemHandler, authClient)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second, // от Slowloris атак
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
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
		return err

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

func newRouter(bookH *handler.BookHandler, reviewH *handler.ReviewHandler, systemH *handler.SystemHandler, tv handler.TokenValidator) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Устанавливаем таймаут
	r.Use(middleware.Timeout(DefaultTimeout))

	// ENDPOINTS
	r.Get("/health", systemH.Health)
	r.Get("/ready", systemH.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		// Публичные
		r.Get("/books", bookH.ListBooks)
		r.Get("/books/{bookId}", bookH.GetBook)

		r.Get("/books/{bookId}/reviews", reviewH.ListBookReviews)
		r.Get("/reviews/{reviewId}", reviewH.GetReview)

		// Защищенные
		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(tv))

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
