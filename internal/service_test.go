package internal

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type MockNanoleafClient struct {
	PairFunc     func(ctx context.Context, ip string) (string, error)
	SetPowerFunc func(ctx context.Context, ip, token string, on bool) error
	GetInfoFunc  func(ctx context.Context, ip, token string) (interface{}, error)
}

func (m *MockNanoleafClient) Pair(ctx context.Context, ip string) (string, error) {
	return m.PairFunc(ctx, ip)
}

func (m *MockNanoleafClient) SetPower(ctx context.Context, ip, token string, on bool) error {
	return m.SetPowerFunc(ctx, ip, token, on)
}

func (m *MockNanoleafClient) GetInfo(ctx context.Context, ip, token string) (interface{}, error) {
	return m.GetInfoFunc(ctx, ip, token)
}

type MockDeviceScanner struct {
	ScanFunc func(ctx context.Context) ([]string, error)
}

func (m *MockDeviceScanner) Scan(ctx context.Context) ([]string, error) {
	return m.ScanFunc(ctx)
}

type MockConfigManager struct {
	SaveFunc   func(ip, token string) error
	LoadFunc   func() (Config, error)
	ExistsFunc func() bool
}

func (m *MockConfigManager) Save(ip, token string) error {
	return m.SaveFunc(ip, token)
}

func (m *MockConfigManager) Load() (Config, error) {
	return m.LoadFunc()
}

func (m *MockConfigManager) Exists() bool {
	return m.ExistsFunc()
}

func TestNewNanoleafService(t *testing.T) {
	client := &MockNanoleafClient{}
	scanner := &MockDeviceScanner{}
	configManager := &MockConfigManager{}

	service := NewNanoleafService(client, scanner, configManager)

	if service == nil {
		t.Error("Expected NewNanoleafService to return a non-nil service")
	}
}

func TestNanoleafService_ScanForDevices(t *testing.T) {
	tests := []struct {
		name            string
		scanDevices     []string
		scanErr         error
		expectedSuccess bool
		expectedMessage string
		expectedData    interface{}
	}{
		{
			name:            "Successful Scan",
			scanDevices:     []string{"192.168.1.100"},
			scanErr:         nil,
			expectedSuccess: true,
			expectedMessage: "Found 1 device(s)",
			expectedData:    []string{"192.168.1.100"},
		},
		{
			name:            "No Devices Found",
			scanDevices:     []string{},
			scanErr:         nil,
			expectedSuccess: false,
			expectedMessage: "No devices detected",
			expectedData:    []string{},
		},
		{
			name:            "Scan Error",
			scanDevices:     nil,
			scanErr:         errors.New("scanner error"),
			expectedSuccess: false,
			expectedMessage: "Scan failed: scanner error",
			expectedData:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := &MockDeviceScanner{
				ScanFunc: func(ctx context.Context) ([]string, error) {
					return tt.scanDevices, tt.scanErr
				},
			}
			service := NewNanoleafService(&MockNanoleafClient{}, scanner, &MockConfigManager{})
			result := service.ScanForDevices(context.Background())

			if result.Success != tt.expectedSuccess {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, result.Success)
			}
			if result.Message != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, result.Message)
			}
			if tt.expectedData != nil || result.Data != nil {
				if fmt.Sprintf("%v", result.Data) != fmt.Sprintf("%v", tt.expectedData) {
					t.Errorf("Expected data %v, got %v", tt.expectedData, result.Data)
				}
			}
		})
	}
}

func TestNanoleafService_PairWithDevice(t *testing.T) {
	tests := []struct {
		name            string
		ip              string
		pairToken       string
		pairErr         error
		saveErr         error
		expectedSuccess bool
		expectedMessage string
	}{
		{
			name:            "Successful Pair",
			ip:              "192.168.1.100",
			pairToken:       "test_token",
			pairErr:         nil,
			saveErr:         nil,
			expectedSuccess: true,
			expectedMessage: "Device paired successfully",
		},
		{
			name:            "No IP Provided",
			ip:              "",
			pairToken:       "",
			pairErr:         nil,
			saveErr:         nil,
			expectedSuccess: false,
			expectedMessage: "No device IP provided. Please scan first.",
		},
		{
			name:            "Pairing Error",
			ip:              "192.168.1.100",
			pairToken:       "",
			pairErr:         errors.New("client pair error"),
			saveErr:         nil,
			expectedSuccess: false,
			expectedMessage: "Pairing failed: client pair error",
		},
		{
			name:            "Save Config Error",
			ip:              "192.168.1.100",
			pairToken:       "test_token",
			pairErr:         nil,
			saveErr:         errors.New("save error"),
			expectedSuccess: true, // Save error doesn't fail the pairing result, but should be logged/handled elsewhere
			expectedMessage: "Device paired successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockNanoleafClient{
				PairFunc: func(ctx context.Context, ip string) (string, error) {
					return tt.pairToken, tt.pairErr
				},
			}
			configManager := &MockConfigManager{
				SaveFunc: func(ip, token string) error {
					return tt.saveErr
				},
			}
			service := NewNanoleafService(client, &MockDeviceScanner{}, configManager)
			result := service.PairWithDevice(context.Background(), tt.ip)

			if result.Success != tt.expectedSuccess {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, result.Success)
			}
			if result.Message != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, result.Message)
			}
		})
	}
}

func TestNanoleafService_SetDevicePower(t *testing.T) {
	tests := []struct {
		name            string
		ip              string
		token           string
		on              bool
		setPowerErr     error
		expectedSuccess bool
		expectedMessage string
	}{
		{
			name:            "Successful Turn On",
			ip:              "192.168.1.100",
			token:           "test_token",
			on:              true,
			setPowerErr:     nil,
			expectedSuccess: true,
			expectedMessage: "Device turned on successfully",
		},
		{
			name:            "Successful Turn Off",
			ip:              "192.168.1.100",
			token:           "test_token",
			on:              false,
			setPowerErr:     nil,
			expectedSuccess: true,
			expectedMessage: "Device turned off successfully",
		},
		{
			name:            "No IP or Token",
			ip:              "",
			token:           "",
			on:              true,
			setPowerErr:     nil,
			expectedSuccess: false,
			expectedMessage: "Device not paired. Please scan and pair first.",
		},
		{
			name:            "Set Power Error",
			ip:              "192.168.1.100",
			token:           "test_token",
			on:              true,
			setPowerErr:     errors.New("client set power error"),
			expectedSuccess: false,
			expectedMessage: "Failed to turn on device: client set power error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockNanoleafClient{
				SetPowerFunc: func(ctx context.Context, ip, token string, on bool) error {
					return tt.setPowerErr
				},
			}
			service := NewNanoleafService(client, &MockDeviceScanner{}, &MockConfigManager{})
			result := service.SetDevicePower(context.Background(), tt.ip, tt.token, tt.on)

			if result.Success != tt.expectedSuccess {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, result.Success)
			}
			if result.Message != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, result.Message)
			}
		})
	}
}

func TestNanoleafService_LoadConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		exists          bool
		loadConfig      Config
		loadErr         error
		expectedSuccess bool
		expectedMessage string
		expectedData    interface{}
	}{
		{
			name:            "Successful Load",
			exists:          true,
			loadConfig:      Config{IP: "192.168.1.100", Token: "loaded_token"},
			loadErr:         nil,
			expectedSuccess: true,
			expectedMessage: "Configuration loaded successfully",
			expectedData:    Config{IP: "192.168.1.100", Token: "loaded_token"},
		},
		{
			name:            "No Saved Configuration",
			exists:          false,
			loadConfig:      Config{},
			loadErr:         nil,
			expectedSuccess: false,
			expectedMessage: "No saved configuration found",
			expectedData:    nil,
		},
		{
			name:            "Load Error",
			exists:          true,
			loadConfig:      Config{},
			loadErr:         errors.New("config load error"),
			expectedSuccess: false,
			expectedMessage: "Failed to load configuration: config load error",
			expectedData:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configManager := &MockConfigManager{
				ExistsFunc: func() bool {
					return tt.exists
				},
				LoadFunc: func() (Config, error) {
					return tt.loadConfig, tt.loadErr
				},
			}
			service := NewNanoleafService(&MockNanoleafClient{}, &MockDeviceScanner{}, configManager)
			result := service.LoadConfiguration()

			if result.Success != tt.expectedSuccess {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, result.Success)
			}
			if result.Message != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, result.Message)
			}
			if fmt.Sprintf("%v", result.Data) != fmt.Sprintf("%v", tt.expectedData) {
				t.Errorf("Expected data %v, got %v", tt.expectedData, result.Data)
			}
		})
	}
}

func TestNanoleafService_GetDeviceInfo(t *testing.T) {
	tests := []struct {
		name            string
		ip              string
		token           string
		getInfoData     interface{}
		getInfoErr      error
		expectedSuccess bool
		expectedMessage string
		expectedData    interface{}
	}{
		{
			name:            "Successful Get Info",
			ip:              "192.168.1.100",
			token:           "test_token",
			getInfoData:     map[string]interface{}{"name": "My Device"},
			getInfoErr:      nil,
			expectedSuccess: true,
			expectedMessage: "Device info retrieved successfully",
			expectedData:    map[string]interface{}{"name": "My Device"},
		},
		{
			name:            "No IP or Token",
			ip:              "",
			token:           "",
			getInfoData:     nil,
			getInfoErr:      nil,
			expectedSuccess: false,
			expectedMessage: "Device not paired. Please scan and pair first.",
		},
		{
			name:            "Get Info Error",
			ip:              "192.168.1.100",
			token:           "test_token",
			getInfoData:     nil,
			getInfoErr:      errors.New("client get info error"),
			expectedSuccess: false,
			expectedMessage: "Failed to get device info: client get info error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockNanoleafClient{
				GetInfoFunc: func(ctx context.Context, ip, token string) (interface{}, error) {
					return tt.getInfoData, tt.getInfoErr
				},
			}
			service := NewNanoleafService(client, &MockDeviceScanner{}, &MockConfigManager{})
			result := service.GetDeviceInfo(context.Background(), tt.ip, tt.token)

			if result.Success != tt.expectedSuccess {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, result.Success)
			}
			if result.Message != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, result.Message)
			}
			if fmt.Sprintf("%v", result.Data) != fmt.Sprintf("%v", tt.expectedData) {
				t.Errorf("Expected data %v, got %v", tt.expectedData, result.Data)
			}
		})
	}
}
