package main

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/moolite/bot/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

var version string = "0.10.0"
var Cfg *config.Config

var (
	flagDebug bool
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if isatty.IsTerminal(os.Stderr.Fd()) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Str("version", version).
		Msg("starting marrano-bot")

	pflag.BoolVarP(&flagDebug, "verbose", "v", false, "set verbose output")
	pflag.Parse()

	if flagDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

}
