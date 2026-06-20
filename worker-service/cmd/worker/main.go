package main

import (
	"bookshelf/pkg/minio"
	"bookshelf/pkg/rabbitmq"
	"bookshelf/worker-service/internal/config"
	"bookshelf/worker-service/internal/database"
	"bookshelf/worker-service/internal/handler"
	applogger "bookshelf/worker-service/internal/logger"
	"bookshelf/worker-service/internal/processor"
	"bookshelf/worker-service/internal/repository"
	"bookshelf/worker-service/internal/service"
	"bookshelf/worker-service/internal/storage/s3"
	app_amqp "bookshelf/worker-service/internal/transport/rabbitmq"
	"context"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	env := "dev"
	logger := applogger.New(env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Пробрасываем логгер в контекст
	ctx = applogger.WithContext(ctx, logger)

	if err := run(ctx, logger); err != nil {
		logger.Error("failed to start worker service", slog.Any("error", err))
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

	consumer, err := app_amqp.NewConsumer(rabbitMQClient)
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bookRepo := repository.NewBookRepository(db)
	coverRepo := repository.NewCoverRepository(db)

	imageProc := processor.NewCoverProcessor(400, 600, 100, 150)
	imageStorage := s3.NewImageStorage(minioClient, cfg.Storage.Bucket)

	coverService := service.NewCoverService(bookRepo, coverRepo, imageStorage, imageProc)
	imageHandler := handler.NewImageHandler(coverService)

	// Регистрируем обработчик
	consumer.RegisterHandler(cfg.Broker.QueueName, imageHandler.HandleImageCompress)

	// Запускаем consumer
	if err := consumer.Start(ctx); err != nil {
		log.Error("consumer start", slog.Any("error", err))
	}
	log.Info("consumer started")

	// Блокируем main goroutine
	<-ctx.Done()
	return nil
}
