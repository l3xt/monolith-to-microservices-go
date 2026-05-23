package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	ErrLoadVersion        = errors.New("failed to load version")
	ErrLoadServerPort     = errors.New("failed to load port value")
	ErrLoadDBUrl          = errors.New("failed to load db url")
	ErrLoadAuthServiceURL = errors.New("failed to load auth service url")
	ErrLoadServiceKey     = errors.New("failed to load service key")
)

type Config struct {
	Version        string
	Port           string
	DatabaseURL    string
	AuthServiceURL string
	ServiceKey     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	version, ok := os.LookupEnv("VERSION")
	if !ok {
		return nil, ErrLoadVersion
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		return nil, ErrLoadServerPort
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

	return &Config{
		Version:        version,
		Port:           port,
		DatabaseURL:    dbUrl,
		AuthServiceURL: authService,
		ServiceKey:     serviceKey,
	}, nil
}
