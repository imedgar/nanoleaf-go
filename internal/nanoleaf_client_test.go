package internal

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type MockHTTPClient struct {
	DoFunc func(req *HTTPRequest) (*HTTPResponse, error)
}

func (m *MockHTTPClient) Do(req *HTTPRequest) (*HTTPResponse, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, errors.New("DoFunc not set")
}

func TestNewAPIClient(t *testing.T) {
	mockClient := &MockHTTPClient{}
	apiClient := NewAPIClient(mockClient)

	if apiClient == nil {
		t.Error("Expected APIClient to be initialized, but it was nil")
	}
}

func TestAPIClient_Pair(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *HTTPResponse
		mockError     error
		expectedToken string
		expectedError bool
	}{
		{
			name: "Successful Pairing",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte(`{"auth_token": "test_token"}`),
			},
			expectedToken: "test_token",
			expectedError: false,
		},
		{
			name:          "HTTP Error",
			mockError:     errors.New("network error"),
			expectedToken: "",
			expectedError: true,
		},
		{
			name: "Non-200 Status Code",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       []byte(`"error": "internal server error"}`),
			},
			expectedToken: "",
			expectedError: true,
		},
		{
			name: "Invalid JSON Response",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte(`invalid json}`),
			},
			expectedToken: "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *HTTPRequest) (*HTTPResponse, error) {
					return tt.mockResponse, tt.mockError
				},
			}
			apiClient := NewAPIClient(mockClient)

			token, err := apiClient.Pair(context.Background(), "192.168.1.100")

			if tt.expectedError {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if token != tt.expectedToken {
					t.Errorf("Expected token %q, but got %q", tt.expectedToken, token)
				}
			}
		})
	}
}

func TestAPIClient_SetPower(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *HTTPResponse
		mockError     error
		expectedError bool
	}{
		{
			name: "Successful SetPower",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusNoContent,
			},
			expectedError: false,
		},
		{
			name:          "HTTP Error",
			mockError:     errors.New("network error"),
			expectedError: true,
		},
		{
			name: "Non-204 Status Code",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       []byte(`"error": "bad request"}`),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *HTTPRequest) (*HTTPResponse, error) {
					if strings.Contains(req.URL, "state") && req.Method == "PUT" {
						var payload map[string]interface{}
						if err := json.Unmarshal(req.Body, &payload); err != nil {
							t.Errorf("Failed to unmarshal request body: %v", err)
						}
						if onVal, ok := payload["on"].(map[string]interface{})["value"].(bool); ok {
							if onVal != true { // Assuming we are testing setting power to true
								t.Errorf("Expected 'on' value to be true, got %v", onVal)
							}
						} else {
							t.Error("Could not find 'on.value' in request body")
						}
					}
					return tt.mockResponse, tt.mockError
				},
			}
			apiClient := NewAPIClient(mockClient)

			err := apiClient.SetPower(context.Background(), "192.168.1.100", "test_token", true)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
			}
		})
	}
}

func TestAPIClient_GetInfo(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *HTTPResponse
		mockError     error
		expectedError bool
		expectedInfo  interface{}
	}{
		{
			name: "Successful GetInfo",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte(`{"name": "My Nanoleaf", "state": {"on": {"value": true}}}`),
			},
			expectedError: false,
			expectedInfo:  map[string]interface{}{"name": "My Nanoleaf", "state": map[string]interface{}{"on": map[string]interface{}{"value": true}}},
		},
		{
			name:          "HTTP Error",
			mockError:     errors.New("network error"),
			expectedError: true,
			expectedInfo:  nil,
		},
		{
			name: "Non-200 Status Code",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusNotFound,
				Body:       []byte(`"error": "not found"}`),
			},
			expectedError: true,
			expectedInfo:  nil,
		},
		{
			name: "Invalid JSON Response",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte(`invalid json}`),
			},
			expectedError: true,
			expectedInfo:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *HTTPRequest) (*HTTPResponse, error) {
					return tt.mockResponse, tt.mockError
				},
			}
			apiClient := NewAPIClient(mockClient)

			info, err := apiClient.GetInfo(context.Background(), "192.168.1.100", "test_token")

			if tt.expectedError {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				marshalledGot, _ := json.Marshal(info)
				marshalledExpected, _ := json.Marshal(tt.expectedInfo)
				if string(marshalledGot) != string(marshalledExpected) {
					t.Errorf("Expected info %v, but got %v", tt.expectedInfo, info)
				}
			}
		})
	}
}

func TestAPIClient_SetBrightness(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *HTTPResponse
		mockError     error
		brightness    int
		expectedError bool
	}{
		{
			name: "Successful SetBrightness",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusNoContent,
			},
			brightness:    50,
			expectedError: false,
		},
		{
			name:          "HTTP Error",
			mockError:     errors.New("network error"),
			brightness:    50,
			expectedError: true,
		},
		{
			name: "Non-204 Status Code",
			mockResponse: &HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       []byte(`"error": "bad request"}`),
			},
			brightness:    50,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *HTTPRequest) (*HTTPResponse, error) {
					if strings.Contains(req.URL, "state") && req.Method == "PUT" {
						var payload map[string]interface{}
						if err := json.Unmarshal(req.Body, &payload); err != nil {
							t.Errorf("Failed to unmarshal request body: %v", err)
						}
						if brightnessVal, ok := payload["brightness"].(map[string]interface{})["value"].(float64); ok {
							if int(brightnessVal) != tt.brightness {
								t.Errorf("Expected brightness value to be %d, got %f", tt.brightness, brightnessVal)
							}
						} else {
							t.Error("Could not find 'brightness.value' in request body")
						}
					}
					return tt.mockResponse, tt.mockError
				},
			}
			apiClient := NewAPIClient(mockClient)

			err := apiClient.SetBrightness(context.Background(), "192.168.1.100", "test_token", tt.brightness)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
			}
		})
	}
}
