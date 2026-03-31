package config

import (
	"log/slog"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestLoadFileEmpty(t *testing.T) {
	is := is.New(t)

	cfg, err := LoadFile("")
	is.NoErr(err)
	is.True(cfg != nil)
	is.Equal(cfg.Database, "")
	is.Equal(cfg.Port, 0)
}

func TestLoadFileValid(t *testing.T) {
	is := is.New(t)

	f, err := os.CreateTemp("", "config-*.toml")
	is.NoErr(err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(`
database = "test.sqlite"
port = 1234

[telegram]
name = "testbot"
domain = "bot.example.com"
token = "123456:abc"
apikey = "secret"
`)
	is.NoErr(err)
	f.Close()

	cfg, err := LoadFile(f.Name())
	is.NoErr(err)
	is.Equal(cfg.Database, "test.sqlite")
	is.Equal(cfg.Port, 1234)
	is.Equal(cfg.Telegram.Name, "testbot")
	is.Equal(cfg.Telegram.Domain, "bot.example.com")
	is.Equal(cfg.Telegram.Token, "123456:abc")
	is.Equal(cfg.Telegram.ApiKey, "secret")
}

func TestLoadFileNotFound(t *testing.T) {
	is := is.New(t)

	_, err := LoadFile("/nonexistent/path/config.toml")
	is.True(err != nil)
}

func TestConcileWithEnvPort(t *testing.T) {
	is := is.New(t)

	os.Setenv("PORT", "9999")
	defer os.Unsetenv("PORT")

	cfg := &Config{}
	concileWithEnv(cfg)
	is.Equal(cfg.Port, 9999)
}

func TestConcileWithEnvDatabase(t *testing.T) {
	is := is.New(t)

	os.Setenv("DATABASE", "env-db.sqlite")
	defer os.Unsetenv("DATABASE")

	cfg := &Config{}
	concileWithEnv(cfg)
	is.Equal(cfg.Database, "env-db.sqlite")
}

func TestConcileWithEnvTelegramToken(t *testing.T) {
	is := is.New(t)

	os.Setenv("TELEGRAM_TOKEN", "env-token")
	defer os.Unsetenv("TELEGRAM_TOKEN")

	cfg := &Config{}
	concileWithEnv(cfg)
	is.Equal(cfg.Telegram.Token, "env-token")
}

func TestConcileWithEnvTelegramKey(t *testing.T) {
	is := is.New(t)

	os.Setenv("TELEGRAM_KEY", "env-key")
	defer os.Unsetenv("TELEGRAM_KEY")

	cfg := &Config{}
	concileWithEnv(cfg)
	is.Equal(cfg.Telegram.ApiKey, "env-key")
}

func TestConcileWithEnvLogLevel(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		env   string
		level slog.Level
	}{
		{"error", slog.LevelError},
		{"warn", slog.LevelWarn},
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}

	for _, tc := range tests {
		t.Run(tc.env, func(t *testing.T) {
			is := is.New(t)
			if tc.env != "" {
				os.Setenv("LOG_LEVEL", tc.env)
			} else {
				os.Unsetenv("LOG_LEVEL")
			}

			cfg := &Config{}
			concileWithEnv(cfg)
			is.Equal(cfg.LogLevel, tc.level)
		})
	}
	os.Unsetenv("LOG_LEVEL")
}

func TestConcileWithEnvOverridesFile(t *testing.T) {
	is := is.New(t)

	os.Setenv("PORT", "8080")
	os.Setenv("DATABASE", "override.sqlite")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("DATABASE")

	f, err := os.CreateTemp("", "config-*.toml")
	is.NoErr(err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(`
database = "file.sqlite"
port = 6446
`)
	is.NoErr(err)
	f.Close()

	cfg, err := LoadFile(f.Name())
	is.NoErr(err)
	is.Equal(cfg.Port, 8080)
	is.Equal(cfg.Database, "override.sqlite")
}

func TestConcileWithEnvInvalidPort(t *testing.T) {
	is := is.New(t)

	os.Setenv("PORT", "not-a-number")
	defer os.Unsetenv("PORT")

	cfg := &Config{Port: 6446}
	concileWithEnv(cfg)
	is.Equal(cfg.Port, 6446)
}
