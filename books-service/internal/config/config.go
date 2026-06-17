package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

const (
	DefaultVersion       = "1.0.0"
	DefaultPort          = "8082"
	DefaultStorageBucket = "books-bucket"
)

var (
	ErrLoadVersion               = errors.New("failed to load version")
	ErrLoadServerPort            = errors.New("failed to load port value")
	ErrLoadDBUrl                 = errors.New("failed to load db url")
	ErrLoadAuthServiceURL        = errors.New("failed to load auth service url")
	ErrLoadServiceKey            = errors.New("failed to load service key")
	ErrLoadStorageEndpoint       = errors.New("failed to load storage endpoint")
	ErrLoadStoragePublicEndpoint = errors.New("failed to load storage public endpoint")
	ErrLoadStorageAccessKey      = errors.New("failed to load storage access key")
	ErrLoadStorageSecretKey      = errors.New("failed to load storage secret key")
	ErrLoadRabbitMQURL           = errors.New("failed to load rabbitmq url")
)

type Config struct {
	Version        string
	Port           string
	DatabaseURL    string
	AuthServiceURL string
	ServiceKey     string
	Storage        StorageConfig
	RabbitMQURL    string
}
type StorageConfig struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	version, ok := os.LookupEnv("VERSION")
	if !ok {
		version = DefaultVersion
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = DefaultPort
	}

	dbUrl, ok := os.LookupEnv("DB_URL")
	if !ok {
		return nil, ErrLoadDBUrl
	}

	authService, ok := os.LookupEnv("AUTH_SERVICE_URL")
	if !ok {
		return nil, ErrLoadAuthServiceURL
	}

	serviceKey, ok := os.LookupEnv("SERVICE_KEY")
	if !ok {
		return nil, ErrLoadServiceKey
	}

	// STORAGE
	storageEndpoint, ok := os.LookupEnv("STORAGE_ENDPOINT")
	if !ok {
		return nil, ErrLoadStorageEndpoint
	}

	storagePublicEndpoint, ok := os.LookupEnv("STORAGE_PUBLIC_ENDPOINT")
	if !ok {
		return nil, ErrLoadStoragePublicEndpoint
	}

	storageAccessKey, ok := os.LookupEnv("STORAGE_ACCESS_KEY")
	if !ok {
		return nil, ErrLoadStorageAccessKey
	}

	storageSecretKey, ok := os.LookupEnv("STORAGE_SECRET_KEY")
	if !ok {
		return nil, ErrLoadStorageSecretKey
	}

	storageBucket, ok := os.LookupEnv("STORAGE_BUCKET")
	if !ok {
		storageBucket = DefaultStorageBucket
	}

	useSSL := false
	if ssl, ok := os.LookupEnv("STORAGE_USE_SSL"); ok && ssl == "true" {
		useSSL = true
	}

	// RabbitMQ
	rabbitmqURL, ok := os.LookupEnv("RABBITMQ_URL")
	if !ok {
		return nil, ErrLoadRabbitMQURL
	}

	return &Config{
		Version:        version,
		Port:           port,
		DatabaseURL:    dbUrl,
		AuthServiceURL: authService,
		ServiceKey:     serviceKey,
		Storage: StorageConfig{
			Endpoint:       storageEndpoint,
			PublicEndpoint: storagePublicEndpoint,
			AccessKey:      storageAccessKey,
			SecretKey:      storageSecretKey,
			Bucket:         storageBucket,
			UseSSL:         useSSL,
		},
		RabbitMQURL: rabbitmqURL,
	}, nil
}
