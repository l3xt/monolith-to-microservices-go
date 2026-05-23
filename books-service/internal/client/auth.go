package client

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/transport/http/dto"
	"bookshelf/pkg/httpclient"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type AuthClient struct {
	http       *httpclient.Client
	serviceKey string
}

func NewAuthClient(hc *httpclient.Client, key string) *AuthClient {
	return &AuthClient{
		http:       hc,
		serviceKey: key,
	}
}

func (c *AuthClient) VerifyToken(ctx context.Context, req *dto.TokenRequest) (*dto.VerifyResponse, error) {
	var resp dto.VerifyResponse

	headers := map[string]string{
		"X-Service-Key": c.serviceKey,
	}

	err := c.http.Post(ctx, "/internal/v1/auth/verify", req, &resp, headers)
	if err != nil {
		// Оборачиваем в нашу доменную ошибку
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrAuthServiceUnavailable, err)
	}
	return &resp, nil
}

func (c *AuthClient) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]dto.UserPublic, error) {
	req := dto.GetUsersRequest{UserIDs: ids}
	var resp dto.GetUsersResponse

	headers := map[string]string{
		"X-Service-Key": c.serviceKey,
	}

	err := c.http.Post(ctx, "/internal/v1/users/batch", req, &resp, headers)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// Оборачиваем в нашу доменную ошибку
		return nil, fmt.Errorf("%w: %v", domain.ErrAuthServiceUnavailable, err)
	}
	return resp.Users, nil
}

func (c *AuthClient) HealthCheck(ctx context.Context) error {
	var resp dto.HealthResponse

	err := c.http.Get(ctx, "/health", &resp)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w: %v", domain.ErrAuthServiceUnavailable, err)
	}
	return nil
}