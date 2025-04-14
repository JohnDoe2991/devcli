package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	filename = "devcli.log"
)

var logger zerolog.Logger
var loggerInit = false

func initLog() {
	runLogFile, err := os.OpenFile(
		filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot open log file")
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05", NoColor: false}
	multi := zerolog.MultiLevelWriter(consoleWriter, runLogFile)
	logger = zerolog.New(multi).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	loggerInit = true
}

func GetLogger(modulename string) zerolog.Logger {
	if !loggerInit {
		initLog()
	}
	return logger.With().Str("module", modulename).Logger()
}

func SetLevelFromString(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		log.Warn().Str("level", level).Msg("unknown log level")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
