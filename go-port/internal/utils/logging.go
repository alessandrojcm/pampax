package utils

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoggingOptions struct {
	Pretty  bool
	Verbose bool
	Writer  io.Writer
}

func SetupLogger(opts LoggingOptions) zerolog.Logger {
	writer := opts.Writer
	if writer == nil {
		writer = os.Stdout
	}

	level := zerolog.InfoLevel
	if opts.Verbose {
		level = zerolog.DebugLevel
	}

	out := writer
	if opts.Pretty {
		out = zerolog.ConsoleWriter{Out: writer, TimeFormat: time.RFC3339}
	}

	logger := zerolog.New(out).Level(level).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(level)
	log.Logger = logger

	return logger
}
