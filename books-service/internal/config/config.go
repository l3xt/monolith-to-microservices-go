package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	ErrLoadServerPort = errors.New("failed to load port value")
	ErrLoadDBUrl      = errors.New("failed to load db url")
	ERrLoadAuthServiceURL = errors.New("failed to load auth service url")
)

type Config struct {
	Port        string
	DatabaseURL string
	AuthServiceURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

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
		return nil, ERrLoadAuthServiceURL
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbUrl,
		AuthServiceURL: authService,
	}, nil
}
