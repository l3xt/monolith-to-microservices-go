package main

import (
	"bookshelf/users-service/internal/config"
	"bookshelf/users-service/internal/database"
	"bookshelf/users-service/internal/handler"
	appjwt "bookshelf/users-service/internal/lib/jwt"
	applogger "bookshelf/users-service/internal/logger"
	"bookshelf/users-service/internal/repository"
	"bookshelf/users-service/internal/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Инициализация инфраструктуры
	jwtManager := appjwt.NewJWTManager(cfg.SecretKey)

	// Инициализация репозиториев
	repo := repository.NewUserRepository(db)

	// Инициализация бизнес-логики
	userService := service.NewUserService(repo)
	authService := service.NewAuthService(userService, jwtManager)

	// Инициализация транспортного слоя
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	systemHandler := handler.NewSystemHandler(db)
	internalHandler := handler.NewInternalHandler(jwtManager, userService)

	router := newRouter(log, authHandler, userHandler, systemHandler, internalHandler, jwtManager, cfg.ServiceKey)

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

func newRouter(logger *slog.Logger, authH *handler.AuthHandler, userH *handler.UserHandler, systemH *handler.SystemHandler, internalH *handler.InternalHandler, tokenManager service.TokenManager, serviceKey string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
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
	r.Get("/ready", systemH.Health)

	r.Route("/api/v1", func(r chi.Router) {
		// Публичные
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/logout", authH.Logout)

		// Защищенные
		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(tokenManager))

			r.Get("/users/me", userH.GetCurrentUser)
			r.Put("/users/me", userH.UpdateCurrentUser)
		})
	})

	r.Route("/internal/v1", func(r chi.Router) {
		r.Use(handler.ServiceKeyMiddleware(serviceKey))

		r.Post("/auth/verify", internalH.VerifyToken)
		r.Post("/users/batch", internalH.GetUsersByIDs)
	})

	return r
}
