package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type HTTPRequest struct {
	Method  string
	URL     string
	Body    []byte
	Headers map[string]string
	Timeout time.Duration
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Status     string
}

type DefaultHTTPClient struct {
	client *http.Client
	mu     sync.RWMutex
}

// NewDefaultHTTPClient creates a new HTTP client with connection pooling
func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     30 * time.Second,
			},
		},
	}
}

// Do executes an HTTP request
func (c *DefaultHTTPClient) Do(req *HTTPRequest) (*HTTPResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if req.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
		defer cancel()
		return c.doWithContext(ctx, req)
	}

	return c.doWithContext(context.Background(), req)
}

func (c *DefaultHTTPClient) doWithContext(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error) {
	var body io.Reader
	if req.Body != nil {
		body = bytes.NewBuffer(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set default content type if not specified
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Status:     resp.Status,
	}, nil
}

// Close closes the HTTP client and cleans up resources
func (c *DefaultHTTPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	return nil
}
