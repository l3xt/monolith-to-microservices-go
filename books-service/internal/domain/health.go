package domain

import (
	"context"
	"errors"
)

var ErrAuthServiceUnavailable = errors.New("authentication service not available")

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}
