package config

import (
	"fmt"
	"log/slog"
	"strings"
)

// LogLevel is a type wrapper for an int type. It indicates the different log level
type LogLevel int

const (
	LevelUnknown LogLevel = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

// UnmarshalString satisfies the fig.StringUnmarshaler interface for the LogLevel type
func (l *LogLevel) UnmarshalString(value string) error {
	switch strings.ToLower(value) {
	case "trace":
		*l = LevelTrace
	case "debug":
		*l = LevelDebug
	case "info":
		*l = LevelInfo
	case "warn":
		*l = LevelWarn
	case "error":
		*l = LevelError
	default:
		return fmt.Errorf("unknown log level: %s", value)
	}
	return nil
}

// String satisfies the fmt.Stringer interface for the LogLevel type
func (l LogLevel) String() string {
	switch l {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// SLog returns the slog.Level value for the LogLevel type
func (l LogLevel) SLog() slog.Level {
	switch l {
	case LevelTrace:
		return slog.LevelDebug
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelUnknown:
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}
