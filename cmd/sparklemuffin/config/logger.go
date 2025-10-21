// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// LogFormatConsole indicates log messages are formatted as pretty-printed key-value pairs.
	LogFormatConsole = "console"

	// LogFormatJSON indicates log messages are formatted as JSON documents.
	LogFormatJSON = "json"
)

var (
	// LogLevelValues lists available logging levels.
	LogLevelValues = []string{
		zerolog.LevelTraceValue,
		zerolog.LevelDebugValue,
		zerolog.LevelInfoValue,
		zerolog.LevelWarnValue,
		zerolog.LevelErrorValue,
		zerolog.LevelFatalValue,
		zerolog.LevelPanicValue,
	}
)

// SetupGlobalLogger configures the global zerolog.Logger.
func SetupGlobalLogger(logFormat string, logLevelValue string) error {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	switch logFormat {
	case LogFormatConsole:
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})

	case LogFormatJSON:
		log.Logger = log.Output(os.Stderr)

	default:
		log.Error().Str("format", logFormat).Msg("log: invalid format")
		return fmt.Errorf("log: invalid format %q", logFormat)
	}

	var logLevel zerolog.Level

	if err := logLevel.UnmarshalText([]byte(logLevelValue)); err != nil {
		log.Error().Err(err).Msg("invalid log level")
		return err
	}

	zerolog.SetGlobalLevel(logLevel)

	return nil
}
