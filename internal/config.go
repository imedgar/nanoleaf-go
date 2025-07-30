package internal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	IP    string `json:"ip"`
	Token string `json:"token"`
}

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".nanoleaf_config.json")
}

func saveConfig(ip, token string) error {
	config := Config{IP: ip, Token: token}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), data, 0600)
}

func loadConfig() (Config, error) {
	var config Config
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func configExists() bool {
	_, err := os.Stat(getConfigPath())
	return !errors.Is(err, os.ErrNotExist)
}
