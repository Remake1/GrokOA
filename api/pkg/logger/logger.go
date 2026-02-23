package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Level           string
	Format          string
	TimeFieldFormat string
	IncludeCaller   bool
}

func New(cfg Config) (zerolog.Logger, error) {
	level := zerolog.InfoLevel
	if cfg.Level != "" {
		parsedLevel, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
		if err != nil {
			return zerolog.Logger{}, fmt.Errorf("parse log level %q: %w", cfg.Level, err)
		}
		level = parsedLevel
	}

	timeFieldFormat := cfg.TimeFieldFormat
	if timeFieldFormat == "" {
		timeFieldFormat = time.RFC3339
	}
	zerolog.TimeFieldFormat = timeFieldFormat

	var writer io.Writer = os.Stdout
	if strings.EqualFold(cfg.Format, "console") {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: timeFieldFormat,
		}
	}

	logger := zerolog.New(writer).With().Timestamp().Logger().Level(level)
	if cfg.IncludeCaller {
		logger = logger.With().Caller().Logger()
	}

	return logger, nil
}
