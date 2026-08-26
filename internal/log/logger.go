package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/wneessen/listenstats/internal/config"
)

type Logger struct {
	*slog.Logger
}

func New(conf *config.Config) *Logger {
	return NewWithReplaceAttr(conf, nil)
}

func NewWithReplaceAttr(conf *config.Config, replaceAttr func(groups []string, a slog.Attr) slog.Attr) *Logger {
	var logger *slog.Logger
	if conf == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		return &Logger{logger}
	}

	level := conf.Log.Level.SLog()
	var output *os.File

	switch strings.ToLower(conf.Log.Output) {
	case "stderr":
		output = os.Stderr
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

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
