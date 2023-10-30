package config

import (
	"os"
	"strconv"

	"github.com/pelletier/go-toml"
)

type TelegramConfig struct {
	Name  string `toml:"name"`
	Token string `toml:"token"`
}

type Config struct {
	Database string         `toml:"database"`
	Port     int            `toml:"port" default:"6446"`
	Telegram TelegramConfig `toml:"telegram"`
}

func concileWithEnv(cfg *Config) {
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}
}

func LoadFromEnv() (*Config, error) {
	filename := os.Getenv("config")
	cfg, err := LoadFile(filename)
	if err != nil {
		return nil, err
	}

	concileWithEnv(cfg)
	return cfg, nil
}

func LoadFile(filename string) (*Config, error) {
	cfg := &Config{}
	if filename != "" {
		tree, err := toml.LoadFile(filename)
		if err != nil {
			return nil, err
		}
		if err = tree.Unmarshal(cfg); err != nil {
			return nil, err
		}
	}

	concileWithEnv(cfg)
	return cfg, nil
}
