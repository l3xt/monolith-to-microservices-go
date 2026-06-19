package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultVersion       = "1.0.0"
	DefaultPort          = "8082"
	DefaultStorageBucket = "books-bucket"

	DefaultDBMaxConns          int32 = 10
	DefaultDBMinConns          int32 = 2
	DefaultDBMaxConnLifetime         = time.Hour
	DefaultDBMaxConnIdleTime         = 5 * time.Minute
	DefaultDBHealthCheckPeriod       = 10 * time.Second
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
	Database       DatabaseConfig
	Storage        StorageConfig
	Broker         BrokerConfig
}

type DatabaseConfig struct {
	URL               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type StorageConfig struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
}

type BrokerConfig struct {
	URL       string
	QueueName string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	version := getEnv("VERSION", DefaultVersion)
	port := getEnv("PORT", DefaultPort)

	authService, ok := os.LookupEnv("AUTH_SERVICE_URL")
	if !ok {
		return nil, ErrLoadAuthServiceURL
	}

	serviceKey, ok := os.LookupEnv("SERVICE_KEY")
	if !ok {
		return nil, ErrLoadServiceKey
	}

	// DATABASE
	dbUrl, ok := os.LookupEnv("DB_URL")
	if !ok {
		return nil, ErrLoadDBUrl
	}

	dbMaxConns := getEnv("DB_MAX_CONNS", DefaultDBMaxConns)
	dbMinConns := getEnv("DB_MIN_CONNS", DefaultDBMinConns)
	dbMaxConnLifetime := getEnv("DB_MAX_CONN_LIFETIME", DefaultDBMaxConnLifetime)
	dbMaxConnIdleTime := getEnv("DB_MAX_CONN_IDLE_TIME", DefaultDBMaxConnIdleTime)
	dbHealthCheckPeriod := getEnv("DB_HEALTH_CHECK_PERIOD", DefaultDBHealthCheckPeriod)

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

	storageBucket := getEnv("STORAGE_BUCKET", DefaultStorageBucket)
	useSSL := getEnv("STORAGE_USE_SSL", false)

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
		Database: DatabaseConfig{
			URL:               dbUrl,
			MaxConns:          dbMaxConns,
			MinConns:          dbMinConns,
			MaxConnLifetime:   dbMaxConnLifetime,
			MaxConnIdleTime:   dbMaxConnIdleTime,
			HealthCheckPeriod: dbHealthCheckPeriod,
		},
		Storage: StorageConfig{
			Endpoint:       storageEndpoint,
			PublicEndpoint: storagePublicEndpoint,
			AccessKey:      storageAccessKey,
			SecretKey:      storageSecretKey,
			Bucket:         storageBucket,
			UseSSL:         useSSL,
		},
		Broker: BrokerConfig{
			URL:       rabbitmqURL,
			QueueName: getEnv("QUEUE_NAME", "image_compress"),
		},
	}, nil
}

func getEnv[T any](key string, defaultValue T) T {
	valueStr, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	var ret any = defaultValue
	switch any(defaultValue).(type) {
	case string:
		ret = valueStr
	case int32:
		if v, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
			ret = int32(v)
		}
	case time.Duration:
		if v, err := time.ParseDuration(valueStr); err == nil {
			ret = v
		}
	case bool:
		if v, err := strconv.ParseBool(valueStr); err == nil {
			ret = v
		}
	}

	return ret.(T)
}
