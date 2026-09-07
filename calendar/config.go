package calendar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "trail")
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func HasConfig() bool {
	_, err := LoadConfig()
	return err == nil
}

func ResolveCredentials() (clientID, clientSecret string, err error) {
	clientID = os.Getenv("OAUTH_CLIENT_ID")
	clientSecret = os.Getenv("OAUTH_CLIENT_SECRET")

	if clientID != "" && clientSecret != "" {
		return
	}

	cfg, cfgErr := LoadConfig()
	if cfgErr != nil {
		return "", "", fmt.Errorf("no credentials found — run 'trail configure' first, or set OAUTH_CLIENT_ID and OAUTH_CLIENT_SECRET environment variables")
	}

	return cfg.ClientID, cfg.ClientSecret, nil
}
