package internal

import (
	"context"
	"fmt"
)

type ConfigManager interface {
	Save(ip, token string) error
	Load() (Config, error)
	Exists() bool
}

type DeviceScanner interface {
	Scan(ctx context.Context) ([]string, error)
}

type NanoleafService struct {
	client        NanoleafClient
	scanner       DeviceScanner
	configManager ConfigManager
}

type ServiceResult struct {
	Success bool
	Message string
	Data    interface{}
}

// NewNanoleafService creates a new Nanoleaf service
func NewNanoleafService(client NanoleafClient, scanner DeviceScanner, configManager ConfigManager) *NanoleafService {
	return &NanoleafService{
		client:        client,
		scanner:       scanner,
		configManager: configManager,
	}
}

// ScanForDevices scans for Nanoleaf devices on the network
func (s *NanoleafService) ScanForDevices(ctx context.Context) ServiceResult {
	devices, err := s.scanner.Scan(ctx)
	if err != nil {
		return ServiceResult{
			Success: false,
			Message: fmt.Sprintf("Scan failed: %s", err.Error()),
			Data:    nil,
		}
	}

	if len(devices) == 0 {
		return ServiceResult{
			Success: false,
			Message: "No devices detected",
			Data:    []string{},
		}
	}

	return ServiceResult{
		Success: true,
		Message: fmt.Sprintf("Found %d device(s)", len(devices)),
		Data:    devices,
	}
}

// PairWithDevice pairs with a Nanoleaf device
func (s *NanoleafService) PairWithDevice(ctx context.Context, ip string) ServiceResult {
	if ip == "" {
		return ServiceResult{
			Success: false,
			Message: "No device IP provided. Please scan first.",
			Data:    nil,
		}
	}

	token, err := s.client.Pair(ctx, ip)
	if err != nil {
		return ServiceResult{
			Success: false,
			Message: fmt.Sprintf("Pairing failed: %s", err.Error()),
			Data:    nil,
		}
	}

	s.configManager.Save(ip, token)
	return ServiceResult{
		Success: true,
		Message: "Device paired successfully",
		Data:    token,
	}
}

// SetDevicePower controls the power state of a Nanoleaf device
func (s *NanoleafService) SetDevicePower(ctx context.Context, ip, token string, on bool) ServiceResult {
	if ip == "" || token == "" {
		return ServiceResult{
			Success: false,
			Message: "Device not paired. Please scan and pair first.",
			Data:    nil,
		}
	}

	action := "off"
	if on {
		action = "on"
	}

	err := s.client.SetPower(ctx, ip, token, on)
	if err != nil {
		return ServiceResult{
			Success: false,
			Message: fmt.Sprintf("Failed to turn %s device: %s", action, err.Error()),
			Data:    nil,
		}
	}

	return ServiceResult{
		Success: true,
		Message: fmt.Sprintf("Device turned %s successfully", action),
		Data:    nil,
	}
}

// LoadConfiguration loads the saved configuration
func (s *NanoleafService) LoadConfiguration() ServiceResult {
	if !s.configManager.Exists() {
		return ServiceResult{
			Success: false,
			Message: "No saved configuration found",
			Data:    nil,
		}
	}

	config, err := s.configManager.Load()
	if err != nil {
		return ServiceResult{
			Success: false,
			Message: fmt.Sprintf("Failed to load configuration: %s", err.Error()),
			Data:    nil,
		}
	}

	return ServiceResult{
		Success: true,
		Message: "Configuration loaded successfully",
		Data:    config,
	}
}

// GetDeviceInfo retrieves information about a device
func (s *NanoleafService) GetDeviceInfo(ctx context.Context, ip, token string) ServiceResult {
	if ip == "" || token == "" {
		return ServiceResult{
			Success: false,
			Message: "Device not paired. Please scan and pair first.",
			Data:    nil,
		}
	}

	info, err := s.client.GetInfo(ctx, ip, token)
	if err != nil {
		return ServiceResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get device info: %s", err.Error()),
			Data:    nil,
		}
	}

	return ServiceResult{
		Success: true,
		Message: "Device info retrieved successfully",
		Data:    info,
	}
}

// Set Brightness for a device
func (s *NanoleafService) SetBrightness(ctx context.Context, ip, token string, b int) ServiceResult {
	if ip == "" || token == "" {
		return ServiceResult{
			Success: false,
			Message: "Device not paired. Please scan and pair first.",
			Data:    nil,
		}
	}

	err := s.client.SetBrightness(ctx, ip, token, b)
	if err != nil {
		return ServiceResult{
			Success: false,
			Message: fmt.Sprintf("Failed to set device brightness: %s", err.Error()),
			Data:    nil,
		}
	}

	return ServiceResult{
		Success: true,
		Message: fmt.Sprintf("Brightness set to %d successfully", b),
		Data:    nil,
	}
}
