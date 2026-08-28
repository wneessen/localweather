package log

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/wneessen/localweather/internal/config"
)

// Logger is a wrapper around slog.Logger that provides extended logging functionality for the application.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger instance based on the provided configuration.
func New(conf *config.Config) *Logger {
	return NewWithReplaceAttr(conf, nil)
}

// NewWithReplaceAttr creates a new Logger instance with a configuration and an optional attribute replacement function.
func NewWithReplaceAttr(conf *config.Config, replaceAttr func(groups []string, a slog.Attr) slog.Attr) *Logger {
	var logger *slog.Logger
	if conf == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		return &Logger{logger}
	}

	level := conf.Log.Level.SLog()
	var output io.Writer

	switch strings.ToLower(conf.Log.Output) {
	case "stderr":
		output = os.Stderr
	case "discard":
		output = io.Discard
	default:
		output = os.Stdout
	}

	switch strings.ToLower(conf.Log.Format) {
	case "json":
		logger = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: replaceAttr,
		}))
	default:
		logger = slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: replaceAttr,
		}))
	}

	return &Logger{logger}
}

// ErrAttr creates a slog.Attr with the key "error" and the given error value.
func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
