package client

import (
	"bookshelf/books-service/internal/transport/http/dto"
	"bookshelf/pkg/httpclient"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Client struct {
	http *httpclient.Client
}

func NewClient(hc *httpclient.Client) *Client {
	return &Client{
		http: hc,
	}
}

func (c *Client) VerifyToken(ctx context.Context, req *dto.TokenRequest) (*dto.VerifyResponse, error) {
	var resp dto.VerifyResponse

	err := c.http.Post(ctx, "/internal/v1/auth/verify", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("auth client: verify failed: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]dto.UserPublic, error) {
	req := dto.GetUsersRequest{UserIDs: ids}
	var resp dto.GetUsersResponse

	err := c.http.Post(ctx, "/internal/v1/users/batch", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("auth client: get users failed: %w", err)
	}
	return resp.Users, nil
}
