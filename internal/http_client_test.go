package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDefaultHTTPClient(t *testing.T) {
	client := NewDefaultHTTPClient()
	if client.client == nil {
		t.Error("Expected http.Client to be initialized, but it was nil")
	}
}

func TestDefaultHTTPClient_Do(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer server.Close()

	client := NewDefaultHTTPClient()
	req := &HTTPRequest{
		Method: "GET",
		URL:    server.URL,
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	expectedBody := "Hello, client\n"
	if string(resp.Body) != expectedBody {
		t.Errorf("Expected body '%s', but got '%s'", expectedBody, string(resp.Body))
	}
}

func TestDefaultHTTPClient_DoWithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprintln(w, "Hello, client")
	}))
	defer server.Close()

	client := NewDefaultHTTPClient()
	req := &HTTPRequest{
		Method:  "GET",
		URL:     server.URL,
		Timeout: 50 * time.Millisecond,
	}

	_, err := client.Do(req)
	if err == nil {
		t.Fatal("Expected a timeout error, but got nil")
	}
}

func TestDefaultHTTPClient_Close(t *testing.T) {
	client := NewDefaultHTTPClient()
	err := client.Close()
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
}

func TestDefaultHTTPClient_doWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer server.Close()

	client := NewDefaultHTTPClient()
	req := &HTTPRequest{
		Method: "GET",
		URL:    server.URL,
	}

	resp, err := client.doWithContext(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}
