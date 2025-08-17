package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type NanoleafClient struct {
	httpClient *http.Client
}

func newClient() *NanoleafClient {
	return &NanoleafClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *NanoleafClient) buildURL(ip, path string) string {
	if ip[0:4] == "http" {
		return fmt.Sprintf("%s/%s", ip, path)
	}
	return fmt.Sprintf("http://%s:16021/%s", ip, path)
}

func (c *NanoleafClient) pair(ctx context.Context, ip string) (string, error) {
	url := c.buildURL(ip, "api/v1/new")

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("pairing request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pairing failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AuthToken string `json:"auth_token"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse pairing response: %w", err)
	}

	return result.AuthToken, nil
}

func (c *NanoleafClient) getInfo(ctx context.Context, ip, token string) (map[string]interface{}, error) {
	url := c.buildURL(ip, fmt.Sprintf("api/v1/%s", token))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get info request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get info failed with status %d", resp.StatusCode)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to parse info response: %w", err)
	}

	return info, nil
}

func (c *NanoleafClient) listEffects(ctx context.Context, ip, token string) ([]string, error) {
	url := c.buildURL(ip, fmt.Sprintf("api/v1/%s/effects/effectsList", token))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get list effects request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get list effects failed with status %d", resp.StatusCode)
	}

	var effects []string
	if err := json.NewDecoder(resp.Body).Decode(&effects); err != nil {
		return nil, fmt.Errorf("failed to parse list effects response: %w", err)
	}

	return effects, nil
}

func (c *NanoleafClient) setPower(ctx context.Context, ip, token string, on bool) error {
	url := c.buildURL(ip, fmt.Sprintf("api/v1/%s/state", token))

	payload := map[string]interface{}{
		"on": map[string]bool{"value": on},
	}

	return c.sendStateUpdate(ctx, url, payload)
}

func (c *NanoleafClient) setBrightness(ctx context.Context, ip, token string, brightness int) error {
	url := c.buildURL(ip, fmt.Sprintf("api/v1/%s/state", token))

	payload := map[string]interface{}{
		"brightness": map[string]int{"value": brightness},
	}

	return c.sendStateUpdate(ctx, url, payload)
}

func (c *NanoleafClient) setEffect(ctx context.Context, ip, token, effect string) error {
	url := c.buildURL(ip, fmt.Sprintf("api/v1/%s/effects", token))

	payload := map[string]interface{}{
		"select": effect,
	}

	return c.sendStateUpdate(ctx, url, payload)
}

func (c *NanoleafClient) sendStateUpdate(ctx context.Context, url string, payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("state update request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("state update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
