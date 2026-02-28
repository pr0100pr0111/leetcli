package config

import (
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Username        string `yaml:"username"`
	RefreshInterval int    `yaml:"refresh_interval"`
}

func Load() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	path := filepath.Join(home, ".leetcli.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if cfg.RefreshInterval == 0 {
		cfg.RefreshInterval = 300
	}

	return cfg, nil
}