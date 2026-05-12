package logging

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
)

type Level int

const (
	LevelUnspecified Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelDPanic
	LevelPanic
	LevelFatal
)

func ParseLevel(value string) (Level, error) {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "", "unspecified":
		return LevelUnspecified, nil
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	case "dpanic":
		return LevelDPanic, nil
	case "panic":
		return LevelPanic, nil
	case "fatal":
		return LevelFatal, nil
	default:
		return LevelUnspecified, fmt.Errorf("invalid log level: %q", value)
	}
}

func (l Level) String() string {
	switch l {
	case LevelUnspecified:
		return "unspecified"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelDPanic:
		return "dpanic"
	case LevelPanic:
		return "panic"
	case LevelFatal:
		return "fatal"
	default:
		return "unspecified"
	}
}

func (l Level) ZapLevel() zapcore.Level {
	switch l {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo, LevelUnspecified:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	case LevelDPanic:
		return zapcore.DPanicLevel
	case LevelPanic:
		return zapcore.PanicLevel
	case LevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
