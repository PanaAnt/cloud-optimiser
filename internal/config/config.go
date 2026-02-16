package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AppConfig struct {
	Mode string `json:"mode"` 
}

var defaultConfig = AppConfig{
	Mode: "mock",
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".cloud-optimiser")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0700)
	}

	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() (AppConfig, error) {
	path, err := getConfigPath()
	if err != nil {
		return defaultConfig, err
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		
		err = SaveConfig(defaultConfig)
		return defaultConfig, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return defaultConfig, err
	}

	return cfg, nil
}

func SaveConfig(cfg AppConfig) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0600)
}

