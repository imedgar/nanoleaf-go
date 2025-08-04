package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewDevice(t *testing.T) {
	device := NewDevice()
	if device == nil {
		t.Fatal("NewDevice should return a non-nil device")
	}
	if device.client == nil {
		t.Fatal("device client should not be nil")
	}
}

func TestLoadConfigSuccess(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	testIP := "192.168.1.100"
	testToken := "test-token"
	err := saveConfig(testIP, testToken)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	device := NewDevice()
	err = device.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig should not fail: %v", err)
	}

	if device.config.IP != testIP {
		t.Errorf("expected IP %s, got %s", testIP, device.config.IP)
	}
	if device.config.Token != testToken {
		t.Errorf("expected Token %s, got %s", testToken, device.config.Token)
	}
}

func TestLoadConfigNoFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	device := NewDevice()
	err := device.LoadConfig()
	if err == nil {
		t.Error("LoadConfig should fail when no config file exists")
	}
}

func TestIsDeviceReadyNoConfig(t *testing.T) {
	device := NewDevice()
	ctx := context.Background()

	if device.IsDeviceReady(ctx) {
		t.Error("device should not be ready without config")
	}
}

func TestIsDeviceReadyWithValidConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{"name": "Test Device"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	device := NewDevice()
	device.config.IP = server.URL
	device.config.Token = "test-token"

	ctx := context.Background()
	if !device.IsDeviceReady(ctx) {
		t.Error("device should be ready with valid config and responding server")
	}
}

func TestSetDevice(t *testing.T) {
	device := NewDevice()
	testIP := "192.168.1.100"

	device.SetDevice(testIP)
	if device.config.IP != testIP {
		t.Errorf("expected IP %s, got %s", testIP, device.config.IP)
	}
}

func TestPairDeviceNoIP(t *testing.T) {
	device := NewDevice()
	ctx := context.Background()

	err := device.PairDevice(ctx)
	if err == nil {
		t.Error("PairDevice should fail when no IP is set")
	}
}

func TestPairDeviceSuccess(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expectedToken := "new-auth-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{"auth_token": expectedToken}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	device := NewDevice()
	device.config.IP = server.URL

	ctx := context.Background()
	err := device.PairDevice(ctx)
	if err != nil {
		t.Fatalf("PairDevice should not fail: %v", err)
	}

	if device.config.Token != expectedToken {
		t.Errorf("expected token %s, got %s", expectedToken, device.config.Token)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("config should be saved: %v", err)
	}
	if config.Token != expectedToken {
		t.Errorf("saved token should be %s, got %s", expectedToken, config.Token)
	}
}

func TestTurnOnOff(t *testing.T) {
	var receivedPowerState bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		if onValue, ok := payload["on"].(map[string]interface{}); ok {
			receivedPowerState = onValue["value"].(bool)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	device := NewDevice()
	device.config.IP = server.URL
	device.config.Token = "test-token"

	ctx := context.Background()

	err := device.TurnOn(ctx)
	if err != nil {
		t.Fatalf("TurnOn should not fail: %v", err)
	}
	if !receivedPowerState {
		t.Error("expected power state to be true")
	}

	err = device.TurnOff(ctx)
	if err != nil {
		t.Fatalf("TurnOff should not fail: %v", err)
	}
	if receivedPowerState {
		t.Error("expected power state to be false")
	}
}

func TestSetBrightnessValid(t *testing.T) {
	var receivedBrightness int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		if brightnessValue, ok := payload["brightness"].(map[string]interface{}); ok {
			receivedBrightness = int(brightnessValue["value"].(float64))
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	device := NewDevice()
	device.config.IP = server.URL
	device.config.Token = "test-token"

	ctx := context.Background()
	expectedBrightness := 50

	err := device.SetBrightness(ctx, expectedBrightness)
	if err != nil {
		t.Fatalf("SetBrightness should not fail: %v", err)
	}
	if receivedBrightness != expectedBrightness {
		t.Errorf("expected brightness %d, got %d", expectedBrightness, receivedBrightness)
	}
}

func TestSetBrightnessInvalid(t *testing.T) {
	device := NewDevice()
	ctx := context.Background()

	err := device.SetBrightness(ctx, -1)
	if err == nil {
		t.Error("SetBrightness should fail with negative value")
	}

	err = device.SetBrightness(ctx, 101)
	if err == nil {
		t.Error("SetBrightness should fail with value > 100")
	}
}

func TestGetDeviceIP(t *testing.T) {
	device := NewDevice()
	testIP := "192.168.1.100"
	device.config.IP = testIP

	if device.GetDeviceIP() != testIP {
		t.Errorf("expected IP %s, got %s", testIP, device.GetDeviceIP())
	}
}

func TestCreateContext(t *testing.T) {
	device := NewDevice()
	ctx, cancel := device.createContext()
	defer cancel()

	if ctx == nil {
		t.Fatal("context should not be nil")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("context should have a deadline")
	}

	expectedDeadline := time.Now().Add(10 * time.Second)
	if deadline.After(expectedDeadline.Add(time.Second)) {
		t.Error("context deadline is too far in the future")
	}
}
