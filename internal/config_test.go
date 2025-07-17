package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileConfigManager(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".nanoleaf_config.json")

	t.Run("NewFileConfigManager", func(t *testing.T) {
		homeDir, _ := os.UserHomeDir()
		expectedPath := filepath.Join(homeDir, ".nanoleaf_config.json")

		manager := NewFileConfigManager()
		if manager.configPath != expectedPath {
			t.Errorf("expected config path %s, got %s", expectedPath, manager.configPath)
		}
	})

	t.Run("Save and Load", func(t *testing.T) {
		manager := &FileConfigManager{configPath: configPath}
		ip := "192.168.1.100"
		token := "test-token"

		err := manager.Save(ip, token)
		if err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		if !manager.Exists() {
			t.Fatal("expected config file to exist, but it doesn't")
		}

		config, err := manager.Load()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if config.IP != ip {
			t.Errorf("expected IP %s, got %s", ip, config.IP)
		}

		if config.Token != token {
			t.Errorf("expected token %s, got %s", token, config.Token)
		}
	})

	t.Run("Save empty config", func(t *testing.T) {
		manager := &FileConfigManager{configPath: configPath}
		err := manager.Save("", "")
		if err == nil {
			t.Fatal("expected error when saving empty config, but got nil")
		}
	})

	t.Run("Load non-existent config", func(t *testing.T) {
		manager := &FileConfigManager{configPath: "/non-existent-path"}
		_, err := manager.Load()
		if err == nil {
			t.Fatal("expected error when loading non-existent config, but got nil")
		}
	})

	t.Run("Load invalid config", func(t *testing.T) {
		manager := &FileConfigManager{configPath: configPath}
		invalidConfig := `{"ip": "192.168.1.100"}`
		err := os.WriteFile(configPath, []byte(invalidConfig), 0600)
		if err != nil {
			t.Fatalf("failed to write invalid config file: %v", err)
		}

		_, err = manager.Load()
		if err == nil {
			t.Fatal("expected error when loading invalid config, but got nil")
		}
	})
}
