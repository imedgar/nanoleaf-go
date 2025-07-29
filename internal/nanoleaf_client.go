package internal

import (
	"context"
	"encoding/json"
	"fmt"
)

type NanoleafClient interface {
	Pair(ctx context.Context, ip string) (string, error)
	SetPower(ctx context.Context, ip, token string, on bool) error
	GetInfo(ctx context.Context, ip, token string) (interface{}, error)
	SetBrightness(ctx context.Context, ip, token string, b int) error
}

type HTTPClient interface {
	Do(req *HTTPRequest) (*HTTPResponse, error)
}

type APIClient struct {
	httpClient HTTPClient
}

// NewAPIClient creates a new APIClient.
func NewAPIClient(httpClient HTTPClient) *APIClient {
	return &APIClient{
		httpClient: httpClient,
	}
}

// Pair requests a new authentication token from the Nanoleaf device.
func (c *APIClient) Pair(ctx context.Context, ip string) (string, error) {
	url := fmt.Sprintf("http://%s:16021/api/v1/new", ip)
	req := &HTTPRequest{
		Method: "POST",
		URL:    url,
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("pairing request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("pairing failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var result struct {
		AuthToken string `json:"auth_token"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal pairing response: %w", err)
	}

	return result.AuthToken, nil
}

// SetPower sets the power state of the Nanoleaf device.
func (c *APIClient) SetPower(ctx context.Context, ip, token string, on bool) error {
	url := fmt.Sprintf("http://%s:16021/api/v1/%s/state", ip, token)

	body, err := json.Marshal(map[string]interface{}{
		"on": map[string]interface{}{
			"value": on,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal power state: %w", err)
	}

	req := &HTTPRequest{
		Method: "PUT",
		URL:    url,
		Body:   body,
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("power request failed: %w", err)
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("power request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	return nil
}

// GetInfo retrieves information about the Nanoleaf device.
func (c *APIClient) GetInfo(ctx context.Context, ip, token string) (interface{}, error) {
	url := fmt.Sprintf("http://%s:16021/api/v1/%s", ip, token)
	req := &HTTPRequest{
		Method: "GET",
		URL:    url,
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getinfo request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("getinfo failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var result interface{}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getinfo response: %w", err)
	}

	return result, nil
}

// SetBrightness sets brightness for the Nanoleaf device.
func (c *APIClient) SetBrightness(ctx context.Context, ip, token string, b int) error {
	url := fmt.Sprintf("http://%s:16021/api/v1/%s/state", ip, token)

	body, err := json.Marshal(map[string]interface{}{
		"brightness": map[string]interface{}{
			"value": b,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to alter brightness: %w", err)
	}

	req := &HTTPRequest{
		Method: "PUT",
		URL:    url,
		Body:   body,
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("brightness request failed: %w", err)
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("brightness request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	return nil
}
