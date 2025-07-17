package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".nanoleaf_config.json"

type Config struct {
	IP    string `json:"ip"`
	Token string `json:"token"`
}

type FileConfigManager struct {
	configPath string
}

// NewFileConfigManager creates a new file-based configuration manager
func NewFileConfigManager() *FileConfigManager {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, configFileName)

	return &FileConfigManager{
		configPath: configPath,
	}
}

// Save stores the configuration to file
func (c *FileConfigManager) Save(ip, token string) error {
	if ip == "" || token == "" {
		return errors.New("ip and token cannot be empty")
	}

	config := Config{
		IP:    ip,
		Token: token,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load reads the configuration from file
func (c *FileConfigManager) Load() (Config, error) {
	var config Config

	if !c.Exists() {
		return config, errors.New("config file does not exist")
	}

	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if config.IP == "" || config.Token == "" {
		return config, errors.New("invalid configuration: missing IP or token")
	}

	return config, nil
}

// Exists checks if the configuration file exists
func (c *FileConfigManager) Exists() bool {
	_, err := os.Stat(c.configPath)
	return !os.IsNotExist(err)
}
