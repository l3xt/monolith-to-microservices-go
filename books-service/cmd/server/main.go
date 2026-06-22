package main

import (
	"bookshelf/books-service/internal/client"
	"bookshelf/books-service/internal/config"
	"bookshelf/books-service/internal/database"
	"bookshelf/books-service/internal/handler"
	applogger "bookshelf/books-service/internal/logger"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"bookshelf/books-service/internal/storage/s3"
	app_amqp "bookshelf/books-service/internal/transport/amqp"
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
	"bookshelf/pkg/minio"
	"bookshelf/pkg/rabbitmq"

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

	// Пробрасываем логгер в контекст
	ctx = applogger.WithContext(ctx, logger)

	if err := run(ctx, logger); err != nil {
		logger.Error("failed to start books service", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.NewPostgresDB(
		ctx,
		cfg.Database.URL,
		cfg.Database.MaxConns,
		cfg.Database.MinConns,
		cfg.Database.MaxConnLifetime,
		cfg.Database.MaxConnIdleTime,
		cfg.Database.HealthCheckPeriod,
	)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}
	defer db.Close()
	log.Info("connected to database")

	minioClient, err := minio.New(minio.Config{
		Endpoint:       cfg.Storage.Endpoint,
		PublicEndpoint: cfg.Storage.PublicEndpoint,
		AccessKey:      cfg.Storage.AccessKey,
		SecretKey:      cfg.Storage.SecretKey,
		UseSSL:         cfg.Storage.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}


	rabbitMQClient, err := rabbitmq.NewRabbitMQClient(cfg.Broker.URL)
	if err != nil {
		return fmt.Errorf("failed to create rabbitmq client: %w", err)
	}
	defer rabbitMQClient.Close()

	// Инициализация репозиториев
	bookRepo := repository.NewBookRepository(db)
	coverRepo := repository.NewCoverRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// Инициализация адаптеров
	eventPublisher := app_amqp.NewProducer(rabbitMQClient)
	imageStorage := s3.NewImageStorage(minioClient, cfg.Storage.Bucket)

	// Инициализация бизнес-логики
	bookService := service.NewBookService(bookRepo)
	coverService := service.NewCoverService(bookRepo, coverRepo, imageStorage, eventPublisher)
	reviewService := service.NewReviewService(bookRepo, reviewRepo)

	baseHTTPClient := httpclient.NewClient(cfg.AuthServiceURL, 5*time.Second)
	authClient := client.NewAuthClient(baseHTTPClient, cfg.ServiceKey)

	// Инициализация транспортного слоя
	bookHandler := handler.NewBookHandler(bookService)
	coverHandler := handler.NewCoverHandler(coverService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	systemHandler := handler.NewSystemHandler(cfg.Version, db, authClient, rabbitMQClient, imageStorage)

	router := newRouter(bookHandler, coverHandler, reviewHandler, systemHandler, authClient)

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

func newRouter(bookH *handler.BookHandler, coverH *handler.CoverHandler, reviewH *handler.ReviewHandler, systemH *handler.SystemHandler, tv handler.TokenValidator) *chi.Mux {
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

		r.Get("/books/{bookId}/cover", coverH.GetBookCover)
		r.Get("/books/{bookId}/cover/status", coverH.GetBookCoverStatus)

		// Защищенные
		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(tv))

			r.Post("/books", bookH.CreateBook)
			r.Put("/books/{bookId}", bookH.UpdateBook)
			r.Delete("/books/{bookId}", bookH.DeleteBook)

			r.Post("/books/{bookId}/reviews", reviewH.CreateReview)
			r.Put("/reviews/{reviewId}", reviewH.UpdateReview)
			r.Delete("/reviews/{reviewId}", reviewH.DeleteReview)

			r.Post("/books/{bookId}/cover", coverH.UploadBookCover)
			r.Delete("/books/{bookId}/cover", coverH.DeleteBookCover)
		})
	})

	return r
}
