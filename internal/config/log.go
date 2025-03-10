package config

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/soa-rs/fit/internal/config/logger"
)

var (
	// once is used to ensure that the logger is initialized only once.
	once sync.Once
)

func SetupLogger() {
	once.Do(func() {
		level := GetEnvOrDefault(EnvBackendLogLevel)
		format := GetEnvOrDefault(EnvBackendLogFormat)
		file := GetEnvOrDefault(EnvBackendLogFile)
		output := GetEnvOrDefault(EnvBackendLogOutput)

		parsedLevel, err := zerolog.ParseLevel(level)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to parse log level")
		}
		zerolog.SetGlobalLevel(parsedLevel)

		if format == "pretty" {
			log.Logger = log.Logger.Output(
				zerolog.ConsoleWriter{Out: os.Stdout},
			)
		} else {
			log.Logger = log.Logger.Output(
				zerolog.NewConsoleWriter(),
			)
		}

		if output == "file" {
			file, err := os.OpenFile(
				file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644,
			)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to open log file")
			}
			log.Logger = log.Logger.Output(file)
		}

		logger.MarkInitialized()
	})
}
