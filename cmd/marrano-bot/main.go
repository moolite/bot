package main

import (
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/moolite/bot/internal/config"
	"github.com/moolite/bot/internal/core"
	"github.com/moolite/bot/internal/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

var version string = "0.10.0"
var Cfg *config.Config

var (
	flagHelp       bool
	flagDebug      bool
	flagConfigPath string
	flagInit       bool
	flagDump       bool
)

func parseLogLevel() {
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))

	if flagDebug {
		logLevel = "debug"
	}

	switch logLevel {
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func main() {
	var err error

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if isatty.IsTerminal(os.Stderr.Fd()) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Str("version", version).
		Msg("starting marrano-bot")

	pflag.BoolVarP(&flagDebug, "verbose", "v", false, "set verbose output")
	pflag.StringVarP(&flagConfigPath, "config", "c", "./marrano-bot.toml", "bot configuration path")
	pflag.BoolVarP(&flagHelp, "help", "h", false, "this message")
	pflag.BoolVarP(&flagInit, "init", "I", false, "initialize the database")
	pflag.BoolVarP(&flagDump, "dump", "D", false, "dump configuration object")
	pflag.Parse()

	if flagDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if flagHelp {
		pflag.Usage()
		os.Exit(0)
		return
	}

	Cfg, err = config.LoadFile(flagConfigPath)
	if err != nil {
		log.Error().Err(err).Msg("error loading config file")
		os.Exit(1)
		return
	}

	err = db.Open(Cfg.Database)
	if err != nil {
		log.Error().Err(err).Msg("error opening db")
	}

	if flagInit {
		err := db.CreateTables()
		if err != nil {
			log.Error().Err(err).Msg("error initializing DB")
			os.Exit(1)
			return
		}
		os.Exit(0)
		return
	}

	if flagDump {
		log.Info().Interface("obj", Cfg).Msg("parsed configuration")
		os.Exit(0)
		return
	}

	err = core.Listen(Cfg)
	if err != nil {
		log.Error().Err(err).Msg("server error")
		os.Exit(2)
	}

	os.Exit(0)
}
