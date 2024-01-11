package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml"
)

type TelegramConfig struct {
	Name   string `toml:"name"`
	Domain string `toml:"domain"`
	Token  string `toml:"token"`
	ApiKey string `toml:"apikey"`
}

type Config struct {
	Database string         `toml:"database"`
	Port     int            `toml:"port" default:"6446"`
	Telegram TelegramConfig `toml:"telegram"`
	LogLevel slog.Level
}

func concileWithEnv(cfg *Config) {
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if database := os.Getenv("DATABASE"); database != "" {
		cfg.Database = database
	}

	if token := os.Getenv("TELEGRAM_TOKEN"); token != "" {
		cfg.Telegram.Token = token
	}

	if apikey := os.Getenv("TELEGRAM_KEY"); apikey != "" {
		cfg.Telegram.ApiKey = apikey
	}

	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level slog.Level
	switch logLevel {
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}
	cfg.LogLevel = level
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
