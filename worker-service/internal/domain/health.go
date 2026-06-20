package domain

import (
	"context"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}
