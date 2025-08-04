package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := newClient()
	if client == nil {
		t.Fatal("newClient should return a non-nil client")
	}
	if client.httpClient == nil {
		t.Fatal("httpClient should not be nil")
	}
	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", client.httpClient.Timeout)
	}
}

func TestPairSuccess(t *testing.T) {
	expectedToken := "test-auth-token-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/new" {
			t.Errorf("expected path /api/v1/new, got %s", r.URL.Path)
		}

		response := map[string]string{"auth_token": expectedToken}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newClient()
	ctx := context.Background()

	token, err := client.pair(ctx, server.URL)
	if err != nil {
		t.Fatalf("pair should not fail: %v", err)
	}
	if token != expectedToken {
		t.Errorf("expected token %s, got %s", expectedToken, token)
	}
}

func TestPairFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("pairing not allowed"))
	}))
	defer server.Close()

	client := newClient()
	ctx := context.Background()

	_, err := client.pair(ctx, server.URL)
	if err == nil {
		t.Error("pair should fail with non-200 status")
	}
}

func TestGetInfoSuccess(t *testing.T) {
	expectedInfo := map[string]interface{}{
		"name": "Test Device",
		"state": map[string]interface{}{
			"on": map[string]bool{"value": true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/test-token" {
			t.Errorf("expected path /api/v1/test-token, got %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(expectedInfo)
	}))
	defer server.Close()

	client := newClient()
	ctx := context.Background()

	info, err := client.getInfo(ctx, server.URL, "test-token")
	if err != nil {
		t.Fatalf("getInfo should not fail: %v", err)
	}

	if info["name"] != expectedInfo["name"] {
		t.Errorf("expected name %s, got %s", expectedInfo["name"], info["name"])
	}
}

func TestSetPowerOn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/test-token/state" {
			t.Errorf("expected path /api/v1/test-token/state, got %s", r.URL.Path)
		}

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		onValue, ok := payload["on"].(map[string]interface{})
		if !ok {
			t.Error("expected 'on' field in payload")
		}
		if onValue["value"] != true {
			t.Error("expected 'on.value' to be true")
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newClient()
	ctx := context.Background()

	err := client.setPower(ctx, server.URL, "test-token", true)
	if err != nil {
		t.Fatalf("setPower should not fail: %v", err)
	}
}

func TestSetBrightness(t *testing.T) {
	expectedBrightness := 75

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		brightnessValue, ok := payload["brightness"].(map[string]interface{})
		if !ok {
			t.Error("expected 'brightness' field in payload")
		}
		if int(brightnessValue["value"].(float64)) != expectedBrightness {
			t.Errorf("expected brightness %d, got %v", expectedBrightness, brightnessValue["value"])
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newClient()
	ctx := context.Background()

	err := client.setBrightness(ctx, server.URL, "test-token", expectedBrightness)
	if err != nil {
		t.Fatalf("setBrightness should not fail: %v", err)
	}
}

func TestSendStateUpdateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := newClient()
	ctx := context.Background()

	err := client.setPower(ctx, server.URL, "test-token", true)
	if err == nil {
		t.Error("setPower should fail with non-204 status")
	}
}
