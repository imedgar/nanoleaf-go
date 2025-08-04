package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path := getConfigPath()
	if path == "" {
		t.Error("config path should not be empty")
	}
	if !filepath.IsAbs(path) {
		t.Error("config path should be absolute")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	testIP := "192.168.1.100"
	testToken := "test-token-123"

	err := saveConfig(testIP, testToken)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if config.IP != testIP {
		t.Errorf("expected IP %s, got %s", testIP, config.IP)
	}
	if config.Token != testToken {
		t.Errorf("expected Token %s, got %s", testToken, config.Token)
	}
}

func TestLoadConfigNotExists(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	_, err := loadConfig()
	if err == nil {
		t.Error("expected error when loading non-existent config")
	}
}

func TestConfigExists(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	if configExists() {
		t.Error("config should not exist initially")
	}

	err := saveConfig("test", "test")
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if !configExists() {
		t.Error("config should exist after saving")
	}
}

func TestSaveConfigInvalidJSON(t *testing.T) {
	config := Config{IP: "test", Token: "test"}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatal("json marshal should not fail for valid config")
	}

	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	err = saveConfig(config.IP, config.Token)
	if err != nil {
		t.Fatalf("save config should not fail: %v", err)
	}

	savedData, err := os.ReadFile(getConfigPath())
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	if string(savedData) != string(data) {
		t.Error("saved data does not match expected JSON")
	}
}
