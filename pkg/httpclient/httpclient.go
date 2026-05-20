package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client представляет собой HTTP-клиент для межсервисного взаимодействия
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient создает новый экземпляр HTTP-клиента с заданным базовым URL и таймаутом
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

// Get выполняет GET-запрос и декодирует JSON-ответ в target
func (c *Client) Get(ctx context.Context, path string, target any) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.do(req, target)
}

// Post выполняет POST-запрос с JSON телом и декодирует ответ в target
func (c *Client) Post(ctx context.Context, path string, body, target any, headers ...map[string]string) error {
	url := c.baseURL + path
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if len(headers) > 0 && headers[0] != nil {
		for k, v := range headers[0] {
			req.Header.Set(k, v)
		}
	}

	return c.do(req, target)
}

// Put выполняет PUT-запрос с JSON телом и декодирует ответ в target
func (c *Client) Put(ctx context.Context, path string, body, target any, headers ...map[string]string) error {
	url := c.baseURL + path
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if len(headers) > 0 && headers[0] != nil {
		for k, v := range headers[0] {
			req.Header.Set(k, v)
		}
	}

	return c.do(req, target)
}

// Delete выполняет DELETE-запрос и декодирует JSON-ответ в target
func (c *Client) Delete(ctx context.Context, path string, target any) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.do(req, target)
}

func (c *Client) do(req *http.Request, target any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, err := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if err != nil {
			bodyStr = "failed to read error body"
		}
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, bodyStr)
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
