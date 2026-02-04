package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"collector-fe-instrumentation/internal/domain"
)

const (
	defaultTimeout = 15 * time.Second
	lokiPushPath   = "/loki/api/v1/push"
)

// Client sends log streams to Grafana Loki.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a Loki client. baseURL is the Loki base URL (e.g. https://loki.elvenobservability.com).
func NewClient(baseURL, token string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Push implements usecase.LokiWriter.
func (c *Client) Push(ctx context.Context, tenantID string, streams []domain.LokiStream) error {
	if len(streams) == 0 {
		return nil
	}
	payload := struct {
		Streams []domain.LokiStream `json:"streams"`
	}{Streams: streams}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := c.baseURL
	if len(url) > 0 && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	url += lokiPushPath

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Scope-OrgID", tenantID)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("loki returned %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
