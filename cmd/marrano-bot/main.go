package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/moolite/bot/internal/config"
	"github.com/moolite/bot/internal/core"
	"github.com/moolite/bot/internal/db"
	"github.com/spf13/pflag"
)

var version string = "0.10.0"
var Cfg *config.Config

var (
	flagHelp            bool
	flagDebug           bool
	flagConfigPath      string
	flagInit            bool
	flagDump            bool
	flagExportDB        bool
	flagExportDBPath    string
	flagSyncMedia       bool
	flagSyncMediaFolder string
)

func setupLogging() {
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))

	if flagDebug {
		logLevel = "debug"
	}

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

	var h slog.Handler

	if isatty.IsTerminal(os.Stderr.Fd()) {
		h = tint.NewHandler(os.Stderr, &tint.Options{
			AddSource: true,
			Level:     level,
		})
	} else {
		h = slog.NewJSONHandler(
			os.Stderr,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     level,
			},
		)
	}

	slog.SetDefault(slog.New(h))
}

func main() {
	var err error

	cwd, err := os.Getwd()
	if err != nil {
		cwd = os.TempDir()
		slog.Error("error selecting current working directory, using temp dir", "err", err, "temp dir", cwd)
	}

	slog.Info("starting marrano-bot", "version", version)

	pflag.BoolVarP(&flagDebug, "verbose", "v", false, "set verbose output")
	pflag.StringVarP(&flagConfigPath, "config", "c", "./marrano-bot.toml", "bot configuration path")
	pflag.BoolVarP(&flagHelp, "help", "h", false, "this message")
	pflag.BoolVarP(&flagInit, "init", "I", false, "initialize the database")
	pflag.BoolVarP(&flagDump, "dump", "D", false, "dump configuration object")
	pflag.BoolVarP(&flagExportDB, "export", "E", false, "export database data as csv (defaults to stdout)")
	pflag.StringVar(&flagExportDBPath, "export-dir", cwd, "folder to write database exported data csv files")
	pflag.StringVarP(&flagSyncMediaFolder, "export-media", "M", "", "sync media files to the specified folder.")
	pflag.Parse()

	setupLogging()

	if flagHelp {
		pflag.Usage()
		os.Exit(0)
		return
	}

	Cfg, err = config.LoadFile(flagConfigPath)
	if err != nil {
		slog.Error("error loading config file", "err", err)
		os.Exit(1)
		return
	}

	err = db.Open(Cfg.Database)
	if err != nil {
		slog.Error("error opening db", "err", err)
	}

	if flagInit {
		err := db.Migrate()
		if err != nil {
			slog.Error("error initializing DB", "err", err)
			os.Exit(1)
			return
		}
		os.Exit(0)
		return
	}

	if flagDump {
		slog.Info("parsed configuration", "cfg", Cfg)
		os.Exit(0)
		return
	}

	if flagExportDB {
		files, err := db.ExportDBToFiles(flagExportDBPath)
		if err != nil {
			slog.Error("error exporting to files", "export path", flagExportDBPath, "err", err)
			return
		}

		for _, f := range files {
			slog.Info("written", "file", f)
		}

		os.Exit(0)
		return
	}

	if flagSyncMediaFolder != "" {
		slog.Info("sync media to folder", "folder", flagSyncMediaFolder)
		if err := SyncFolder(flagSyncMediaFolder); err != nil {
			slog.Error("error syncronizing media folder", "folder", flagSyncMediaFolder, "err", err)
			os.Exit(1)
		}
		os.Exit(0)
		return
	}

	err = core.Listen(Cfg)
	if err != nil {
		slog.Error("server error", "err", err)
		os.Exit(2)
	}

	os.Exit(0)
}
